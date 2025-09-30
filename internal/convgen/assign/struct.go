package assign

import (
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"maps"
	"slices"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/convgen/match"
	"github.com/sublee/convgen/internal/convgen/parse"
	"github.com/sublee/convgen/internal/typeinfo"
)

// structAssigner assigns a struct type to another struct type by matching
// fields and methods.
type structAssigner struct {
	x, y    Object // must be struct types
	matches []matchAssigner[structField]
	errWrap *errWrapAssigner
}

// requiresErr returns true if any of the matches has an error.
func (as structAssigner) requiresErr() bool {
	for _, pair := range as.matches {
		if pair.requiresErr() {
			return true
		}
		if pair.X.getter != nil && pair.X.getter.HasErr() {
			return true
		}
		if pair.Y.setter != nil && pair.Y.setter.HasErr() {
			return true
		}
	}
	return false
}

// tryStruct tries to create a [structAssigner] from x to y by matching fields
// and methods.
func (fac *factory) tryStruct(x, y Object) (*structAssigner, error) {
	if !x.Type().IsStruct() || !y.Type().IsStruct() {
		return nil, skip
	}

	m := match.NewMatcher[structField](fac.inj, fac.cfg, x, y)
	errs := discover(fac, m, structDiscovery{
		cfg: fac.cfg,
		pkg: fac.Pkg(),
		x:   x,
		y:   y,
	})
	matches, err := m.Match()
	errs = errors.Join(errs, err)

	// Check if getters and setters return an error.
	if !fac.allowsErr {
		needErr := func(fn typeinfo.Func) error {
			return codefmt.Errorf(fac, fac.inj, `cannot return error of %o
	%b: %o (%t)
	try convgen.StructErr`,
				fn, fn,
				fn, fn)
		}

		for _, pair := range matches {
			x, y := pair.X, pair.Y
			if x.getter != nil && x.getter.HasErr() {
				errs = errors.Join(errs, needErr(x.getter))
			}
			if y.setter != nil && y.setter.HasErr() {
				errs = errors.Join(errs, needErr(y.setter))
			}
		}
	}

	matchAssigners, err := buildMatchAssigners(fac, matches)
	errs = errors.Join(errs, err)
	if errs != nil {
		return nil, errs
	}

	return &structAssigner{
		x:       x,
		y:       y,
		matches: matchAssigners,
		errWrap: fac.newErrWrap(),
	}, nil
}

// structField is a pair of an input field and an output field.
type structField struct {
	// owner is a struct object which this field belongs to.
	owner Object

	// One of the following must be set:
	field  *types.Var    // regular field
	getter typeinfo.Func // getter method
	setter typeinfo.Func // setter method

	name string
	typ  typeinfo.Type
	pkg  *packages.Package
}

// Type returns the type of the field, the return type of the getter, or the
// parameter type of the setter.
func (o structField) Type() typeinfo.Type { return o.typ }

// QualName returns a qualified name of the field or method, like "User.Name".
func (o structField) QualName() string {
	return codefmt.Sprintf(codefmt.Pkg(o.pkg), "%q.%s", o.owner, o.name)
}

// CrumbName returns the crumb name, a dot-separated path of nested field names
// for the field or method. For example, "Session.SignedUser.Name" where
// "Session" is the root struct and "SignedUser" is a field of "Session".
// The type of "Session.SignedUser" is User, but a crumb name records field
// names rather than type names.
func (o structField) CrumbName() string {
	return codefmt.Sprintf(codefmt.Pkg(o.pkg), "%s.%s", o.owner.CrumbName(), o.name)
}

// CrumbNameAfter returns the crumb name, like [structField.CrumbName], but
// relative to the given base object. This is needed to build the correct field
// path in [structAssigner.writeFlattenCode], because subconverters take fields
// from the root object, not from the nested one.
func (o structField) CrumbNameAfter(base Object) string {
	if o.owner == base {
		return o.name
	}
	s, ok := o.owner.(structField)
	if !ok {
		return o.name
	}
	return s.CrumbNameAfter(base) + "." + o.name
}

// DebugName returns the crumb name with its type for debugging. For example,
// "Session.SignedUser.Name (string)".
func (o structField) DebugName() string {
	return codefmt.Sprintf(codefmt.Pkg(o.pkg), "%s (%t)", o.CrumbName(), o.Type())
}

// Exported returns true if the field or method is exported.
func (o structField) Exported() bool {
	if o.field != nil {
		return o.field.Exported()
	}
	if o.getter != nil {
		return o.getter.Object().Exported()
	}
	if o.setter != nil {
		return o.setter.Object().Exported()
	}
	return false
}

// Pkg returns the package of the field or method. It is used for indexing when
// matching.
func (o structField) Pos() token.Pos {
	if o.field != nil {
		return o.field.Pos()
	}
	if o.getter != nil {
		return o.getter.Object().Pos()
	}
	if o.setter != nil {
		return o.setter.Object().Pos()
	}
	return token.NoPos
}

// structDiscovery discovers fields and getter/setter methods of struct types.
type structDiscovery struct {
	pkg  *packages.Package
	cfg  parse.Config
	x, y Object
}

// DiscoverX discovers fields and getter methods of struct X and nested fields
// if enabled.
func (d structDiscovery) DiscoverX(add addFunc[structField], del deleteFunc) error {
	d.discoverFields(d.x, add)
	d.discoverGetters(d.x, add)

	var errs error
	for _, path := range d.cfg.DiscoverNestedX {
		field, _, err := d.ResolveX(path)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		del(field.Pos())
		d.discoverFields(field, add)
		d.discoverGetters(field, add)
	}
	return errs
}

// DiscoverY discovers fields and setter methods of struct Y and nested fields
// if enabled.
func (d structDiscovery) DiscoverY(add addFunc[structField], del deleteFunc) error {
	d.discoverFields(d.y, add)
	d.discoverSetters(d.y, add)

	var errs error
	for _, path := range d.cfg.DiscoverNestedY {
		field, _, err := d.ResolveY(path)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		del(field.Pos())
		d.discoverFields(field, add)
		d.discoverSetters(field, add)
	}
	return errs
}

// discoverFields discovers fields of the given struct object and adds them.
func (d structDiscovery) discoverFields(owner Object, add addFunc[structField]) {
	for f := range owner.Type().Struct.Fields() {
		field := structField{
			owner: owner,
			field: f,
			name:  f.Name(),
			typ:   typeinfo.TypeOf(f.Type()),
			pkg:   d.pkg,
		}
		add(field, f.Name())
	}
}

// discoverGetters discovers getter methods of the given struct object and adds
// them if enabled.
func (d structDiscovery) discoverGetters(owner Object, add addFunc[structField]) {
	if !d.cfg.DiscoverGettersEnabled {
		return
	}

	t := owner.Type()
	if !t.IsNamed() {
		return
	}

	for m := range t.Named.Methods() {
		if !strings.HasPrefix(m.Name(), d.cfg.DiscoverGettersPrefix) || !strings.HasSuffix(m.Name(), d.cfg.DiscoverGettersSuffix) {
			continue
		}

		fn, err := typeinfo.FuncOf[typeinfo.OnlyY](m)
		if err != nil {
			continue
		}

		field := structField{
			owner:  d.x,
			getter: fn,
			name:   m.Name(),
			typ:    fn.Y(),
			pkg:    d.pkg,
		}
		key := strings.TrimSuffix(strings.TrimPrefix(m.Name(), d.cfg.DiscoverGettersPrefix), d.cfg.DiscoverGettersSuffix)
		add(field, key)
	}
}

// discoverSetters discovers setter methods of the given struct object and adds
// them if enabled.
func (d structDiscovery) discoverSetters(owner Object, add addFunc[structField]) {
	if !d.cfg.DiscoverSettersEnabled {
		return
	}

	t := owner.Type()
	if !t.IsNamed() {
		return
	}

	for m := range t.Named.Methods() {
		if !strings.HasPrefix(m.Name(), d.cfg.DiscoverSettersPrefix) || !strings.HasSuffix(m.Name(), d.cfg.DiscoverSettersSuffix) {
			continue
		}

		fn, err := typeinfo.FuncOf[typeinfo.OnlyX](m)
		if err != nil {
			continue
		}

		field := structField{
			owner:  d.y,
			setter: fn,
			name:   m.Name(),
			typ:    fn.X(),
			pkg:    d.pkg,
		}
		key := strings.TrimSuffix(strings.TrimPrefix(m.Name(), d.cfg.DiscoverSettersPrefix), d.cfg.DiscoverSettersSuffix)
		add(field, key)
	}
}

// ResolveX resolves a crumb to a struct field or getter method of struct X.
func (d structDiscovery) ResolveX(path parse.Path) (structField, string, error) {
	parent, err := d.resolveParent(d.x, path)
	if err != nil {
		return structField{}, "", err
	}

	last := path.StructField[len(path.StructField)-1]
	if field, ok := last.(*types.Var); ok {
		return structField{
			owner: parent,
			field: field,
			name:  field.Name(),
			typ:   typeinfo.TypeOf(field.Type()),
			pkg:   d.pkg,
		}, field.Name(), nil
	}

	if method, ok := last.(*types.Func); ok {
		fn, err := typeinfo.FuncOf[typeinfo.OnlyY](method)
		if err != nil {
			return structField{}, "", err
		}

		key := strings.TrimSuffix(strings.TrimPrefix(fn.Name(), d.cfg.DiscoverGettersPrefix), d.cfg.DiscoverGettersSuffix)

		return structField{
			owner:  parent,
			getter: fn,
			name:   fn.Name(),
			typ:    fn.Y(),
			pkg:    d.pkg,
		}, key, nil
	}

	return structField{}, "", fmt.Errorf("%s is not a getter method", last)
}

// ResolveY resolves a crumb to a struct field or setter method of struct Y.
func (d structDiscovery) ResolveY(path parse.Path) (structField, string, error) {
	parent, err := d.resolveParent(d.y, path)
	if err != nil {
		return structField{}, "", err
	}

	last := path.StructField[len(path.StructField)-1]
	if field, ok := last.(*types.Var); ok {
		return structField{
			owner: parent,
			field: field,
			name:  field.Name(),
			typ:   typeinfo.TypeOf(field.Type()),
			pkg:   d.pkg,
		}, field.Name(), nil
	}

	if method, ok := last.(*types.Func); ok {
		fn, err := typeinfo.FuncOf[typeinfo.OnlyX](method)
		if err != nil {
			return structField{}, "", err
		}

		key := strings.TrimSuffix(strings.TrimPrefix(fn.Name(), d.cfg.DiscoverSettersPrefix), d.cfg.DiscoverSettersSuffix)

		return structField{
			owner:  parent,
			setter: fn,
			name:   fn.Name(),
			typ:    fn.X(),
			pkg:    d.pkg,
		}, key, nil
	}

	return structField{}, "", fmt.Errorf("%s is not a getter method", last)
}

func (d structDiscovery) resolveParent(owner Object, path parse.Path) (Object, error) {
	if len(path.StructField) == 0 {
		panic("empty path")
	}
	if path.StructField[0].Pos() != owner.Pos() {
		panic("not a field of owner")
	}

	parent := owner
	for _, obj := range path.StructField[1 : len(path.StructField)-1] {
		field, ok := obj.(*types.Var)
		if !ok {
			panic("intermediate path must be a field")
		}

		parent = structField{
			owner: parent,
			field: field,
			name:  field.Name(),
			typ:   typeinfo.TypeOf(field.Type()),
			pkg:   d.pkg,
		}
	}
	return parent, nil
}

type matchGroup struct {
	PrefixX, PrefixY string
	Matches          []matchAssigner[structField]
}

// groupMatches returns the matches grouped by the prefix of the field names.
// "" is treated as the special prefix that matches all fields without a prefix.
// This is used to ensure that fields without a prefix are assigned first before
// fields with a prefix.
func (as structAssigner) groupMatches() []matchGroup {
	// Group matches by the prefix of the field names.
	prefixed := make(map[[2]string][]matchAssigner[structField])
	for _, pair := range as.matches {
		pathX := pair.X.CrumbNameAfter(as.x)
		pathY := pair.Y.CrumbNameAfter(as.y)

		var prefixX, prefixY string
		if i := strings.LastIndex(pathX, "."); i != -1 {
			prefixX = pathX[:i]
		}
		if j := strings.LastIndex(pathY, "."); j != -1 {
			prefixY = pathY[:j]
		}

		prefixes := [2]string{prefixX, prefixY}
		prefixed[prefixes] = append(prefixed[prefixes], pair)
	}

	// Sort by prefix names.
	keys := slices.Collect(maps.Keys(prefixed))
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})

	out := make([]matchGroup, 0, len(prefixed))
	for _, k := range keys {
		out = append(out, matchGroup{
			PrefixX: k[0],
			PrefixY: k[1],
			Matches: prefixed[k],
		})
	}
	return out
}

// writeAssignCode writes code that assigns struct x to struct y.
func (as structAssigner) writeAssignCode(w *codefmt.Writer, varX, varY, varErr string) {
	labelEnd := w.Name("end")

	for _, m := range as.groupMatches() {
		as.writeMatchesCode(w, m.Matches, m.PrefixX, m.PrefixY, varX, varY, varErr, labelEnd)
	}

	if varErr != "" {
		w.Printf("goto %s\n", labelEnd)
		w.Printf("%s:\n", labelEnd)
		w.Printf("if %s != nil { %s = *new(%t) }\n", varErr, varY, as.y)
	}
}

// writeMatchesCode writes field assignment codes for the given matches. If the
// field has a prefix, it writes a code to flatten the prefix together.
func (as structAssigner) writeMatchesCode(w *codefmt.Writer, matches []matchAssigner[structField], prefixX, prefixY, varX, varY, varErr, labelEnd string) {
	var pathX, pathY []string
	if prefixX != "" {
		pathX = strings.Split(prefixX, ".")
	}
	if prefixY != "" {
		pathY = strings.Split(prefixY, ".")
	}

	// Intro
	if len(pathX) != 0 && len(pathY) != 0 {
		w.Printf("// (nested) %s.%s -> (nested) %s.%s\n", as.x.QualName(), prefixX, as.y.QualName(), prefixY)
		w.Printf("{\n")
		defer w.Printf("}\n")
	} else if len(pathX) != 0 {
		w.Printf("// (nested) %s.%s -> (flat) %s\n", as.x.QualName(), prefixX, as.y.QualName())
		w.Printf("{\n")
		defer w.Printf("}\n")
	} else if len(pathY) != 0 {
		w.Printf("// (flat) %s -> (nested) %s.%s\n", as.x.QualName(), as.y.QualName(), prefixY)
		w.Printf("{\n")
		defer w.Printf("}\n")
	}

	var next func(x, y typeinfo.Type, cursorX, cursorY int)
	next = func(x, y typeinfo.Type, cursorX, cursorY int) {
		// Write blocks to flatten nested X first because if there are no values
		// to convert, we don't need to proceed assigning Y fields.
		if len(pathX) != 0 && cursorX-1 < len(pathX) {
			if x.IsStruct() && cursorX < len(pathX) {
				f, _ := x.StructField(pathX[cursorX])
				next(typeinfo.TypeOf(f.Type()), y, cursorX+1, cursorY)
			} else if x.IsPointer() {
				// Unwrap pointer to the nested X field.
				w.Printf("if %s.%s != nil {\n", varX, strings.Join(pathX[:cursorX], "."))
				next(*x.Elem, y, cursorX, cursorY)
				w.Printf("}\n")
			}
		}

		// Then write blocks to flatten nested Y fields.
		if len(pathY) != 0 && cursorY-1 < len(pathY) {
			if y.IsStruct() && cursorY < len(pathY) {
				f, _ := y.StructField(pathY[cursorY])
				next(x, typeinfo.TypeOf(f.Type()), cursorX, cursorY+1)
			} else if y.IsPointer() {
				// Assign a value at the pointer to the nested Y field.
				varTmpY := w.Name(strings.Join(append([]string{varY}, pathY[:cursorY]...), "."))
				w.Printf("var %s %t\n", varTmpY, y.Elem)
				w.Printf("%s.%s = &%s\n", varY, strings.Join(pathY[:cursorY], "."), varTmpY)
				next(x, *y.Elem, cursorX, cursorY)
			}
		}
	}
	next(as.x.Type(), as.y.Type(), 0, 0)

	// We have reached the end of both prefixes, so we can write the field
	// assignments.
	for _, m := range matches {
		as.writeFieldAssignCode(w, m, varX, varY, varErr, labelEnd)
	}
}

// writeFieldAssignCode writes code to assign a field X to a field Y.
func (as structAssigner) writeFieldAssignCode(w *codefmt.Writer, m matchAssigner[structField], varX, varY, varErr, labelEnd string) {
	if m.X.setter != nil {
		panic("fieldX cannot be setter")
	}
	if m.Y.getter != nil {
		panic("fieldY cannot be getter")
	}

	gotoEndIfErr := func(varTmpErr string, wrapsErr bool) {
		w.Printf("if %s != nil {\n", varTmpErr)
		if wrapsErr {
			as.errWrap.writeWrapCode(w, varTmpErr)
		}
		w.Printf("%s = %s\n", varErr, varTmpErr)
		w.Printf("goto %s }\n", labelEnd)
	}

	// Comment
	w.Printf("// %s -> %s\n", m.X.QualName(), m.Y.QualName())
	w.Printf("{\n")

	// Get X field
	var varFieldX string
	if m.X.field != nil {
		varFieldX = fmt.Sprintf("%s.%s", varX, m.X.CrumbNameAfter(as.x))
	} else {
		if m.X.getter.HasErr() {
			// TODO: safe?
			varFieldX = w.Name("x" + m.X.CrumbNameAfter(as.x))
			varTmpErr := w.Name("err")
			w.Printf("%s, %s := %s.%s()\n", varFieldX, varTmpErr, varX, m.X.CrumbNameAfter(as.x))
			gotoEndIfErr(varTmpErr, true)
		} else {
			varFieldX = fmt.Sprintf("%s.%s()", varX, m.X.CrumbNameAfter(as.x))
		}
	}

	var varFieldY string
	if m.Y.field != nil {
		varFieldY = fmt.Sprintf("%s.%s", varY, m.Y.CrumbNameAfter(as.y))
	} else {
		varFieldY = w.Name("y" + m.Y.CrumbNameAfter(as.y))
		w.Printf("var %s %t\n", varFieldY, m.Y)
	}

	// Convert X to Y field
	varTmpErr := w.Name("err")
	if m.requiresErr() {
		w.Printf("var %s error\n", varTmpErr)
	}
	m.writeAssignCode(w, varFieldX, varFieldY, varTmpErr)
	if m.requiresErr() {
		gotoEndIfErr(varTmpErr, false)
	}

	// Set Y field
	if m.Y.setter != nil {
		if m.Y.setter.HasErr() {
			varTmpErr := w.Name("err")
			w.Printf("%s := %s.%s(%s)\n", varTmpErr, varY, m.Y.CrumbNameAfter(as.y), varFieldY)
			gotoEndIfErr(varTmpErr, true)
		} else {
			w.Printf("%s.%s(%s)\n", varY, m.Y.CrumbNameAfter(as.y), varFieldY)
		}
	}

	w.Printf("}\n")
}
