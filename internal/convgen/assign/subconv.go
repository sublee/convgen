package assign

import (
	"go/token"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/convgen/parse"
	"github.com/sublee/convgen/internal/typeinfo"
)

// subconv represents an implicit struct-to-struct converter, called
// subconverter. It is generated automatically by Convgen without an explicit
// convgen.Struct or convgen.StructErr directive. The generated code is emitted
// as a standalone function that can be invoked by other generated functions.
type subconv struct {
	*funcAssigner
	conv *conv[*structAssigner]
}

// Implement [Conv] by delegating to the underlying convAssigner.
func (s *subconv) WriteDefineCode(w *codefmt.Writer) { s.conv.WriteDefineCode(w) }
func (s *subconv) Pos() token.Pos                    { return token.NoPos }

// newSubconvLookup creates a lookup table for subconverters from a slice of
// [Conv]. The slice should contain only subconverters created by this package.
func newSubconvLookup(subconvs []Conv) *typeinfo.Lookup[*subconv] {
	lookup := typeinfo.NewLookup[*subconv]()
	for _, conv := range subconvs {
		subconv, ok := conv.(*subconv)
		if !ok {
			panic("subconv slice contains non-subconv")
		}
		lookup.Put(subconv)
	}
	return lookup
}

// trySubconvFunc tries to call an existing subconverter defined in this factory
// or its parent factories.
func (fac *factory) trySubconvFunc(x, y Object) (*funcAssigner, error) {
	for ; fac != nil; fac = fac.parent {
		fn, ok := fac.newSubconvs.Get(x.Type(), y.Type())
		if !ok {
			fn, ok = fac.oldSubconvs.Get(x.Type(), y.Type())
			if !ok {
				continue
			}
		}

		as, err := fac.callFunc(x, y, fn)
		if err == nil {
			return as, nil
		}

		// The subconverter was found but cannot be called because it is
		// incompatible with the current factory. To report the reason,
		// reproduce the same subconverter creation using this factory.
		_, err2 := fac.tryStruct(x, y)
		if err2 == nil {
			panic("callFunc failed but trySubconv succeeded") // should not happen
		}
		return nil, err2
	}
	return nil, skip
}

// exportSubconvs returns subconverters newly defined by this factory as a slice
// of [Conv] sorted by name.
func (fac *factory) exportSubconvs() []Conv {
	var export []Conv
	toName := make(map[Conv]string)

	for subconv := range fac.newSubconvs.Range() {
		export = append(export, subconv)
		toName[subconv] = subconv.Name()
	}

	slices.SortFunc(export, func(i, j Conv) int {
		if toName[i] < toName[j] {
			return -1
		}
		return 1
	})
	return export
}

// forkForSubconv creates a new factory for building subconverters. To commit
// the subconverters to the original factory, call [factory.joinForSubconv] with
// the forked factory.
func (fac *factory) forkForSubconv(allowsErr bool) *factory {
	return &factory{
		inj:         fac.inj,
		cfg:         fac.inj.Module.Config.ForkForStruct(),
		ns:          fac.ns,
		allowsErr:   allowsErr,
		parent:      fac,
		newSubconvs: newSubconvLookup(nil),
	}
}

// joinForSubconv merges subconverters defined in the forked factory into this
// factory.
func (fac *factory) joinForSubconv(forked *factory) {
	for subconv := range forked.newSubconvs.Range() {
		fac.newSubconvs.Put(subconv)
	}
}

// newSubconvName generates a new unique name for a subconverter function.
func (fac *factory) newSubconvName(x, y typeinfo.Type) string {
	return fac.ns.Name(formatSubconvName(fac.inj.Pkg(), x, y, fac.inj.Module))
}

func formatSubconvName(pkg *packages.Package, x, y typeinfo.Type, mod *parse.Module) string {
	var b strings.Builder
	b.WriteString("convgen_")

	if mod.Name != "" {
		b.WriteString(mod.Name)
		b.WriteString("_")
	}

	if x.Pkg() != nil && x.Pkg() != pkg.Types {
		b.WriteString(x.Pkg().Name())
		b.WriteString("_")
	}

	if x.Deref().IsNamed() {
		b.WriteString(x.Deref().Named.Obj().Name())
		b.WriteString("_")
	} else {
		b.WriteString("anon_")
	}

	if y.Pkg() != nil && y.Pkg() != pkg.Types {
		b.WriteString(y.Pkg().Name())
		b.WriteString("_")
	}

	if y.Deref().IsNamed() {
		b.WriteString(y.Deref().Named.Obj().Name())
	} else {
		b.WriteString("anon")
	}
	return b.String()
}

// trySubconv tries to create a [subconvAssigner] from x to y. To make the
// assigner callable by other factories of the same module, call
// [factory.commitSubconvs] later.
func (fac *factory) trySubconv(x, y Object) (*subconv, error) {
	name := fac.newSubconvName(x.Type(), y.Type())

	try := func(fac *factory, name string, x, y Object) (*subconv, error) {
		fn := typeinfo.NewFunc(fac.inj.Pkg().Types, name, x.Type(), y.Type(), fac.allowsErr, true)

		call, _ := fac.callFunc(x, y, fn)
		subconv := &subconv{funcAssigner: call}
		fac.newSubconvs.Put(subconv)

		as, err := fac.tryStruct(x, y)
		if err != nil {
			fac.newSubconvs.Del(x.Type(), y.Type())
			return nil, err
		}

		if !fac.allowsErr && as.requiresErr() {
			fac.newSubconvs.Del(x.Type(), y.Type())
			return nil, skip
		}

		subconv.conv = &conv[*structAssigner]{
			Func:     fn,
			assigner: as,
			pkg:      fac.inj.Pkg(),
			pos:      token.NoPos,
		}

		return subconv, nil
	}

	newFac := fac.forkForSubconv(false)
	subconv, err := try(newFac, name, x, y)
	if err != nil {
		if fac.allowsErr {
			newFac = fac.forkForSubconv(true)
			subconv, err = try(newFac, name, x, y)
		}
		if err != nil {
			return nil, err
		}
	}

	fac.joinForSubconv(newFac)
	return subconv, err
}
