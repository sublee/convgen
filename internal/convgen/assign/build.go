package assign

import (
	"errors"

	"golang.org/x/tools/go/packages"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/convgen/match"
	"github.com/sublee/convgen/internal/convgen/parse"
	"github.com/sublee/convgen/internal/typeinfo"
)

// assigner writes an inline code to assign X to Y.
type assigner interface {
	// requiresErr returns true if the assignment has an error return.
	requiresErr() bool

	// writeAssignCode writes an inline code to assign X to Y. varX, varY, and
	// varErr is the name of the variables allocated for X, Y, and error. The
	// code should update varY and varErr for the assignment result instead of
	// allocating new variables.
	writeAssignCode(w *codefmt.Writer, varX, varY, varErr string)
}

// factory is a builder of assigners. It holds build context information.
type factory struct {
	// inj is the injector this factory targets.
	inj parse.Injector

	// cfg is the configuration for building assigners. For an explicit
	// converter it is the configuration of the injector itself, and for an
	// implicit subconverter it is the configuration of the module where the
	// injector is defined.
	cfg parse.Config

	// ns is the namespace to claim unique identifiers.
	ns codefmt.NS

	// allowsErr indicates whether assigners built by this factory are allowed
	// to return an error. If false, assigners that requires to return an error
	// will fail to be built.
	allowsErr bool

	// parent is the parent factory that this factory is forked from.
	// [factory.forkForSubconv] will set this field.
	parent *factory

	// newSubconvs is a lookup table for implicit converters defined by this
	// factory.
	newSubconvs *typeinfo.Lookup[*subconv]

	// oldSubconvs is a lookup table for implicit converters defined by the
	// previous factory. It is used to avoid duplicating implicit converters
	// which were already defined by the previous factory.
	oldSubconvs *typeinfo.Lookup[*subconv]
}

// Pkg implements [codefmt.Pkger].
func (fac *factory) Pkg() *packages.Package { return fac.inj.Pkg() }

// Build creates an explicit converter for the given injector and implicit
// converters which the explicit converter depends on. This is the main entry
// point of this package.
func Build(inj parse.Injector, ns codefmt.NS, oldSubconvs []Conv) (Conv, []Conv, error) {
	fac := factory{
		inj:         inj,
		cfg:         inj.Config,
		ns:          ns,
		allowsErr:   inj.HasErr(),
		newSubconvs: newSubconvLookup(nil),
		oldSubconvs: newSubconvLookup(oldSubconvs),
	}

	as, err := fac.buildExplicit()
	if err != nil {
		return nil, nil, err
	}

	c := &conv[assigner]{
		Func:     inj,
		assigner: as,
		pkg:      inj.Pkg(),
		pos:      inj.Pos(),
		doc:      inj.Doc,
		comment:  inj.Comment,
	}
	return c, fac.exportSubconvs(), nil
}

// buildExplicit builds an assigner of the explicit converter defined by the
// injector of this factory.
func (fac *factory) buildExplicit() (assigner, error) {
	x, y := typeOnly(fac.inj, fac.inj.X()), typeOnly(fac.inj, fac.inj.Y())
	switch {
	case fac.inj.Struct:
		// convgen.Struct or convgen.StructErr
		if as, err := fac.tryStruct(x, y); !errors.Is(err, skip) {
			return as, err
		}

		// convgen.Struct allows pointer of struct
		if as, err := fac.tryStructPointer(x, y); !errors.Is(err, skip) {
			return as, err
		}

		return nil, codefmt.Errorf(fac.inj, fac.inj, "no struct")

	case fac.inj.Union:
		// convgen.Union or convgen.UnionErr
		if as, err := fac.tryUnion(x, y); !errors.Is(err, skip) {
			return as, err
		}
		return nil, codefmt.Errorf(fac.inj, fac.inj, "no union")

	case fac.inj.Enum:
		// convgen.Enum or convgen.EnumErr
		if as, err := fac.tryEnum(x, y, fac.inj.EnumUnknown); !errors.Is(err, skip) {
			return as, err
		}
		return nil, codefmt.Errorf(fac.inj, fac.inj, "no enum")
	}

	return nil, codefmt.Errorf(fac.inj, fac.inj, "neither struct, union, nor enum")
}

// build builds an assigner to convert x to y which is derived from the injector
// of this factory.
func (fac *factory) build(x, y Object) (assigner, error) {
	// Existing functions
	if as, err := fac.tryMatchFunc(x, y); !errors.Is(err, skip) {
		return as, err
	}
	if as, err := fac.tryModuleFunc(x, y); !errors.Is(err, skip) {
		return as, err
	}

	// Implicit subconverter
	if as, err := fac.trySubconvFunc(x, y); !errors.Is(err, skip) {
		return as, err
	}
	if as, err := fac.trySubconv(x, y); !errors.Is(err, skip) {
		return as, err
	}

	// Primitive types
	if as, err := fac.tryPointer(x, y); !errors.Is(err, skip) {
		return as, err
	}
	if as, err := fac.tryBasic(x, y); !errors.Is(err, skip) {
		return as, err
	}
	if as, err := fac.tryIndex(x, y); !errors.Is(err, skip) {
		return as, err
	}
	if as, err := fac.tryKey(x, y); !errors.Is(err, skip) {
		return as, err
	}

	return nil, codefmt.Errorf(fac, fac.inj, `cannot convert %s to %s
	consider convgen.ImportFunc(func(%t) %t) for explicit conversion`,
		x.DebugName(), y.DebugName(), x, y)
}

// skip is returned by tryXXX methods to indicate that the method cannot handle
// the given types.
var skip = errors.New("skip")

// matchAssigner holds a match along with an assigner. It is used for some
// matches which may have their own assigner to write an inline code to assign X
// to Y.
type matchAssigner[T any] struct {
	match.Match[T]
	assigner
}

// buildMatchAssigners builds assigners for the given matches using the given
// factory. It returns pairs of match and assigner.
func buildMatchAssigners[T Object](fac *factory, matches []match.Match[T]) ([]matchAssigner[T], error) {
	// NOTE: This funtion could be a method of factory, but it is not possible
	// because of generic.
	var errs error
	var assigners []matchAssigner[T]
	for _, m := range matches {
		as, err := fac.build(m.X, m.Y)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		assigners = append(assigners, matchAssigner[T]{
			Match:    m,
			assigner: as,
		})
	}
	return assigners, errs
}
