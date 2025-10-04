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

// enumAssigner assigns an enum type to another enum type by matching enum
// members.
type enumAssigner struct {
	x, y    Object
	unknown *types.Const
	pairs   [][2]enumMember
	errWrap *errWrapAssigner
}

// requiresErr always returns false.
func (as enumAssigner) requiresErr() bool { return false }

// tryEnum tries to create an [enumAssigner] from x to y by matching enum
// members.
func (fac *factory) tryEnum(x, y Object, unknown *types.Const) (*enumAssigner, error) {
	if !x.Type().IsBasic() || !y.Type().IsBasic() {
		return nil, skip
	}

	if !types.Identical(unknown.Type(), y.Type().Type()) {
		return nil, codefmt.Errorf(fac, unknown, "unknown must be of type %t, but got %t", y, unknown)
	}

	m := match.NewMatcher[enumMember](fac.inj, fac.cfg, x, y)
	errs := discover(fac, m, enumDiscovery{
		cfg: fac.cfg,
		pkg: fac.Pkg(),
		x:   x,
		y:   y,
	})
	m.SetUnknownY(unknown.Pos())

	matches, err := m.Match()
	errs = errors.Join(errs, err)
	if err != nil {
		return nil, errs
	}

	pairs := make([][2]enumMember, len(matches))
	for i, pair := range matches {
		pairs[i][0] = pair.X
		pairs[i][1] = pair.Y
	}
	return &enumAssigner{
		x:       x,
		y:       y,
		unknown: unknown,
		pairs:   pairs,
		errWrap: fac.newErrWrap(),
	}, nil
}

type enumMember struct {
	con *types.Const
	typ typeinfo.Type
	pkg *packages.Package
	pos token.Pos
}

func (o enumMember) Type() typeinfo.Type    { return o.typ }
func (o enumMember) QualName() string       { return codefmt.FormatObj(o, o.con) }
func (o enumMember) CrumbName() string      { return codefmt.FormatObj(o, o.con) }
func (o enumMember) DebugName() string      { return codefmt.FormatObj(o, o.con) }
func (o enumMember) Exported() bool         { return o.con.Exported() }
func (o enumMember) Pkg() *packages.Package { return o.pkg }
func (o enumMember) Pos() token.Pos         { return o.pos }

type enumDiscovery struct {
	cfg  parse.Config
	pkg  *packages.Package
	x, y Object
}

func (d enumDiscovery) DiscoverX(add addFunc[enumMember], del deleteFunc) error {
	d.discover(d.x, d.cfg.DiscoverBySamplePkgX.Scope(), add)
	return nil
}

func (d enumDiscovery) DiscoverY(add addFunc[enumMember], del deleteFunc) error {
	d.discover(d.y, d.cfg.DiscoverBySamplePkgY.Scope(), add)
	return nil
}

func (d enumDiscovery) discover(enum Object, scope *types.Scope, add addFunc[enumMember]) {
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)

		con, ok := obj.(*types.Const)
		if !ok {
			continue
		}

		if !types.Identical(con.Type(), enum.Type().Type()) {
			continue
		}

		add(enumMember{
			con: con,
			typ: enum.Type(),
			pkg: d.pkg,
			pos: con.Pos(),
		}, con.Name())
	}
}

func (d enumDiscovery) ResolveX(path parse.Path) (enumMember, string, error) {
	return d.resolve(d.x, path)
}

func (d enumDiscovery) ResolveY(path parse.Path) (enumMember, string, error) {
	return d.resolve(d.y, path)
}

func (d enumDiscovery) resolve(enum Object, path parse.Path) (enumMember, string, error) {
	if path.EnumMember == nil {
		panic("enum member not set")
	}

	return enumMember{
		con: path.EnumMember,
		typ: enum.Type(),
		pkg: d.pkg,
		pos: path.Pos,
	}, path.EnumMember.Name(), nil
}

// writeAssignCode writes code that assigns X to Y by enum member matching.
func (as enumAssigner) writeAssignCode(w *codefmt.Writer, varX, varY, varErr string) {
	printErr := func() {
		w.Printf("%s = %o\n", varY, as.unknown)
		if varErr != "" {
			varConvgenErrors := w.Import("github.com/sublee/convgen/pkg/convgenerrors", "convgenerrors")
			varFmt := w.Import("fmt", "fmt")
			w.Printf("%s = %s.Wrap(\"%s\", %s.Errorf(\"unknown enum member %%v: %%w\", %s, %s.ErrNoMatch))\n",
				varErr, varConvgenErrors,
				as.x.QualName(), varFmt,
				varX, varConvgenErrors)
			as.errWrap.writeWrapCode(w, varErr)
		}
	}

	if len(as.pairs) == 0 {
		printErr()
		return
	}

	// Switch cases for each enum member
	w.Printf("switch %s {\n", varX)
	for _, pair := range as.pairs {
		w.Printf("case %o:\n", pair[0].con)
		w.Printf("%s = %o\n", varY, pair[1].con)
		if varErr != "" {
			w.Printf("%s = nil\n", varErr)
		}
	}

	w.Printf("default:\n")
	printErr()
	w.Printf("}\n")
}
