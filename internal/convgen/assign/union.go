package assign

import (
	"errors"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/convgen/match"
	"github.com/sublee/convgen/internal/convgen/parse"
	"github.com/sublee/convgen/internal/typeinfo"
)

// unionAssigner assigns an interface to another interface by matching their
// implementations.
type unionAssigner struct {
	x, y    Object // must be interfaces
	matches []matchAssigner[unionImpl]
	errWrap *errWrapAssigner
}

// requiresErr returns true if any of the matches has an error.
func (as unionAssigner) requiresErr() bool {
	for _, m := range as.matches {
		if m.requiresErr() {
			return true
		}
	}
	return false
}

// tryUnion tries to create a [unionAssigner] from x to y by matching their
// implementations.
func (fac *factory) tryUnion(x, y Object) (*unionAssigner, error) {
	if !x.Type().IsInterface() || !y.Type().IsInterface() {
		return nil, skip
	}

	var errs error
	if x.Type().Interface.NumMethods() == 0 {
		err := codefmt.Errorf(fac, fac.inj, "%t has no methods; union requires at least one", x)
		errs = errors.Join(errs, err)
	}
	if y.Type().Interface.NumMethods() == 0 {
		err := codefmt.Errorf(fac, fac.inj, "%t has no methods; union requires at least one", y)
		errs = errors.Join(errs, err)
	}
	if errs != nil {
		return nil, errs
	}

	m := match.NewMatcher[unionImpl](fac.inj, fac.cfg, x, y)
	errs = discover(fac, m, unionDiscovery{
		cfg: fac.cfg,
		pkg: fac.Pkg(),
		x:   x,
		y:   y,
	})

	matches, err := m.Match()
	errs = errors.Join(errs, err)

	matchAssigners, err := buildMatchAssigners(fac, matches)
	errs = errors.Join(errs, err)

	if errs != nil {
		return nil, errs
	}

	return &unionAssigner{
		x:       x,
		y:       y,
		matches: matchAssigners,
		errWrap: fac.newErrWrap(),
	}, nil
}

// unionImpl is a pair of input and output interface implementations to be
// converted.
type unionImpl struct {
	impl typeinfo.Type
	pkg  *packages.Package
	pos  token.Pos
}

func (o unionImpl) Type() typeinfo.Type    { return o.impl }
func (o unionImpl) QualName() string       { return codefmt.FormatType(o, o.impl.Type()) }
func (o unionImpl) CrumbName() string      { return codefmt.FormatType(o, o.impl.Type()) }
func (o unionImpl) DebugName() string      { return codefmt.FormatType(o, o.impl.Type()) }
func (o unionImpl) Exported() bool         { return o.impl.Deref().Named.Obj().Exported() }
func (o unionImpl) Pkg() *packages.Package { return o.pkg }
func (o unionImpl) Pos() token.Pos         { return o.pos }

type unionDiscovery struct {
	cfg  parse.Config
	pkg  *packages.Package
	x, y Object
}

func (d unionDiscovery) DiscoverX(add addFunc[unionImpl], del deleteFunc) error {
	scope := d.pkg.Types.Scope()
	if d.cfg.DiscoverBySampleEnabled && d.cfg.DiscoverBySamplePkgX != nil {
		scope = d.cfg.DiscoverBySamplePkgX.Scope()
	}
	d.discover(d.x.Type().Interface, scope, add)
	return nil
}

func (d unionDiscovery) DiscoverY(add addFunc[unionImpl], del deleteFunc) error {
	scope := d.pkg.Types.Scope()
	if d.cfg.DiscoverBySampleEnabled && d.cfg.DiscoverBySamplePkgY != nil {
		scope = d.cfg.DiscoverBySamplePkgY.Scope()
	}
	d.discover(d.y.Type().Interface, scope, add)
	return nil
}

func (d unionDiscovery) discover(union *types.Interface, scope *types.Scope, add addFunc[unionImpl]) {
	for _, name := range scope.Names() {
		t := typeinfo.TypeOf(scope.Lookup(name).Type())
		if t.IsInterface() {
			continue
		}
		if t.IsGeneric() {
			continue
		}
		if !t.IsNamed() {
			continue
		}

		if types.AssertableTo(union, t.Named) {
			add(unionImpl{
				impl: t,
				pkg:  d.pkg,
				pos:  t.Pos(),
			}, t.Named.Obj().Name())
		} else if types.AssertableTo(union, t.Ref().Pointer) {
			add(unionImpl{
				impl: t.Ref(),
				pkg:  d.pkg,
				pos:  t.Pos(),
			}, t.Named.Obj().Name())
		}
	}
}

func (d unionDiscovery) ResolveX(path parse.Path) (unionImpl, string, error) {
	return d.resolve(d.x.Type().Interface, path)
}

func (d unionDiscovery) ResolveY(path parse.Path) (unionImpl, string, error) {
	return d.resolve(d.y.Type().Interface, path)
}

func (d unionDiscovery) resolve(union *types.Interface, path parse.Path) (unionImpl, string, error) {
	if path.UnionImpl == nil {
		panic("union impl not set")
	}

	t := typeinfo.TypeOf(path.UnionImpl)
	if t.IsInterface() {
		panic("union impl must not be an interface")
	}
	if t.IsGeneric() {
		panic("union impl must not be a generic type")
	}

	if t.IsNamed() && types.AssertableTo(union, t.Named) {
		// Value receiver implements the interface
		return unionImpl{
			impl: t,
			pkg:  d.pkg,
			pos:  path.Pos,
		}, t.Named.Obj().Name(), nil
	}

	if t.IsPointer() && t.Elem.IsNamed() && types.AssertableTo(union, t.Pointer) {
		// Pointer receiver implements the interface
		return unionImpl{
			impl: t,
			pkg:  d.pkg,
			pos:  path.Pos,
		}, t.Elem.Named.Obj().Name(), nil
	}

	panic("union impl does not implement the interface")
}

// writeAssignCode writes code that assigns interface x to interface y.
func (as unionAssigner) writeAssignCode(w *codefmt.Writer, varX, varY, varErr string) {
	printErr := func() {
		w.Printf("%s = nil\n", varY)
		if varErr != "" {
			varConvgenErrors := w.Import("github.com/sublee/convgen/pkg/convgenerrors", "convgenerrors")
			varFmt := w.Import("fmt", "fmt")
			w.Printf("%s = %s.Wrap(\"%s\", %s.Errorf(\"unknown type %%T\", %s))\n", varErr, varConvgenErrors, as.x.QualName(), varFmt, varX)
		}
	}

	if len(as.matches) == 0 {
		printErr()
		return
	}

	w.Printf("switch %s := %s.(type) {\n", varX, varX)
	for _, m := range as.matches {
		w.Printf("case %t:\n", m.X.impl)

		varTypedY := w.Name("out")
		w.Printf("var %s %t\n", varTypedY, m.Y.impl)

		m.writeAssignCode(w, varX, varTypedY, varErr)
		w.Printf("%s = %s\n", varY, varTypedY)

		if varErr != "" {
			w.Printf("%s = nil\n", varErr)
			as.errWrap.writeWrapCode(w, varErr)
		}
	}

	w.Printf("default:\n")
	printErr()
	w.Printf("}\n")
}
