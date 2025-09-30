package assign

import (
	"go/types"

	"github.com/sublee/convgen/internal/codefmt"
)

// basicAssigner assigns a basic type to another basic type.
//
//	y = x    // for assignable
//	y = T(x) // for convertible without precision loss
type basicAssigner struct {
	assignable  bool
	convertible bool
	x, y        Object
}

// requiresErr always returns false. Basic type conversion never fail as an error.
func (basicAssigner) requiresErr() bool { return false }

// tryBasic tries to create a [basicAssigner] from x to y by either direct
// assignment or type conversion.
func (fac *factory) tryBasic(x, y Object) (*basicAssigner, error) {
	if !x.Type().IsBasic() || !y.Type().IsBasic() {
		// Both x and y must be basic types
		return nil, skip
	}

	if types.AssignableTo(x.Type().Type(), y.Type().Type()) {
		// Can assign x to y directly
		// e.g., y = x
		return &basicAssigner{assignable: true}, nil
	}

	if types.ConvertibleTo(x.Type().Type(), y.Type().Type()) {
		// Can assign x to y with type conversion
		// e.g., y = T(x)
		if err := fac.checkConvertible(x, y); err != nil {
			return nil, err
		}
		return &basicAssigner{convertible: true, x: x, y: y}, nil
	}

	return nil, skip
}

func (fac *factory) checkConvertible(x, y Object) error {
	infoX, infoY := x.Type().Basic.Info(), y.Type().Basic.Info()
	kindX, kindY := x.Type().Basic.Kind(), y.Type().Basic.Kind()

	if infoX&(types.IsInteger|types.IsUnsigned) != 0 && infoY == types.IsString {
		// int or uint -> string conversion is not allowed because it may lose
		// data for non-ASCII characters.
		return skip
	}

	if infoX&(infoX^types.IsUnsigned) != infoY&(infoY^types.IsUnsigned) {
		// Inter-kind conversion is lossy. For example, float to int.
		goto Error
	}

	if infoX&types.IsUnsigned != 0 && infoY&types.IsUnsigned == 0 {
		// uint -> int conversion is allowed only when uint size <= half of int size.
		// For example, uint8 -> int16 is allowed, but uint8 -> int8 is not allowed.
		if kindSizeOf(kindX) > kindSizeOf(kindY)/2 {
			goto Error
		}
	}

	if kindSizeOf(kindX) <= kindSizeOf(kindY) {
		return nil
	}

Error:
	return codefmt.Errorf(fac, fac.inj, `narrowing from %s to %s causes precision loss
	consider convgen.ImportFunc(func(%t) %t) for explicit conversion`,
		x.DebugName(), y.DebugName(), x, y)
}

func kindSizeOf(kind types.BasicKind) int {
	switch kind {
	case types.Bool:
		return 1
	case types.Int8, types.Uint8:
		return 8
	case types.Int16, types.Uint16:
		return 16
	case types.Int, types.Int32, types.Uint, types.Uint32, types.Float32:
		return 32
	case types.Int64, types.Uint64, types.Float64, types.Complex64:
		return 64
	case types.Complex128:
		return 128
	}
	return 999
}

func (as basicAssigner) writeAssignCode(w *codefmt.Writer, varX, varY, varErr string) {
	switch {
	case as.assignable:
		w.Printf("%s = %s\n", varY, varX)
	case as.convertible:
		w.Printf("%s = %t(%s)\n", varY, as.y, varX)
	}
}
