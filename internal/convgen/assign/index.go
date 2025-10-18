package assign

import (
	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/typeinfo"
)

// indexAssigner assigns a slice or array to another slice or array by
// converting each element.
type indexAssigner struct {
	assigner
	elemX, elemY typeinfo.Type
	isSliceY     bool
}

// tryIndex tries to create an [indexAssigner] from x to y by converting each
// element.
func (fac *factory) tryIndex(x, y Object) (*indexAssigner, error) {
	switch {
	case x.Type().IsSlice() && y.Type().IsSlice():
		// Slice to slice
	case x.Type().IsArray() && y.Type().IsSlice():
		// Array to slice
	case x.Type().IsArray() && y.Type().IsArray() && x.Type().Len == y.Type().Len:
		// Array to array with the same length
	default:
		return nil, skip
	}

	elemX := *x.Type().Elem
	elemY := *y.Type().Elem
	as, err := fac.build(typeOnly(fac.inj, elemX), typeOnly(fac.inj, elemY))
	if err != nil {
		return nil, err
	}

	return &indexAssigner{
		assigner: as,
		elemX:    elemX,
		elemY:    elemY,
		isSliceY: y.Type().IsSlice(),
	}, nil
}

// writeAssignCode writes code that assigns x to y by converting each element.
func (a indexAssigner) writeAssignCode(w *codefmt.Writer, varX, varY, varErr string) {
	if a.isSliceY {
		w.Printf("if len(%s) != 0 {\n", varX)
		defer w.Printf("}\n")

		w.Printf("%s = make([]%t, len(%s))\n", varY, a.elemY, varX)
	}

	varI := w.Name("i")
	varV := w.Name("v")
	w.Printf("for %s, %s := range %s {\n", varI, varV, varX)
	defer w.Printf("}\n")

	a.assigner.writeAssignCode(w, varV, varY+"["+varI+"]", varErr)
	if a.requiresErr() {
		w.Printf("if %s != nil {\n", varErr)
		if a.isSliceY {
			w.Printf("%s = nil\n", varX)
		}
		w.Printf("break\n")
		w.Printf("}\n")
	}
}
