package assign

import (
	"errors"
	"fmt"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/typeinfo"
)

// pointerAssigner assigns a pointer to another pointer by unwrapping the
// pointers and converting the underlying element.
type pointerAssigner struct {
	assigner
	elemX, elemY   typeinfo.Type
	depthX, depthY int
}

// tryPointer tries to create a [pointerAssigner] from x to y by unwrapping the
// pointers and converting the underlying element.
func (fac *factory) tryPointer(x, y Object) (*pointerAssigner, error) {
	if !x.Type().IsPointer() && !y.Type().IsPointer() {
		// Either x or y must be a pointer
		return nil, skip
	}

	elemX, elemY := x, y
	if x.Type().IsPointer() {
		elemX = elemOf(x)
	}
	if y.Type().IsPointer() {
		elemY = elemOf(y)
	}

	as, err := fac.build(elemX, elemY)
	if err != nil {
		return nil, err
	}

	return &pointerAssigner{
		assigner: as,
		elemX:    elemX.Type(),
		elemY:    elemY.Type(),
		depthX:   x.Type().PointerDepth(),
		depthY:   y.Type().PointerDepth(),
	}, nil
}

// tryStructPointer tries to create a [pointerAssigner] from x to y by unwrapping the
// pointers and converting the underlying element.
func (fac *factory) tryStructPointer(x, y Object) (*pointerAssigner, error) {
	if !x.Type().IsPointer() && !y.Type().IsPointer() {
		// Either x or y must be a pointer
		return nil, skip
	}

	elemX, elemY := x, y
	if x.Type().IsPointer() {
		elemX = elemOf(x)
	}
	if y.Type().IsPointer() {
		elemY = elemOf(y)
	}

	var as assigner
	as, err := fac.tryStruct(elemX, elemY)
	if err != nil {
		if !errors.Is(err, skip) {
			return nil, err
		}

		as, err = fac.tryStructPointer(elemX, elemY)
		if err != nil {
			return nil, err
		}
	}

	return &pointerAssigner{
		assigner: as,
		elemX:    elemX.Type(),
		elemY:    elemY.Type(),
		depthX:   x.Type().PointerDepth(),
		depthY:   y.Type().PointerDepth(),
	}, nil
}

// writeAssignCode writes code that unwraps the pointers and assigns x to y.
func (as pointerAssigner) writeAssignCode(w *codefmt.Writer, varX, varY, varErr string) {
	if as.depthX != 0 {
		// TODO: nil as zero value
		w.Printf("if %s != nil {\n", varX)
		varX = fmt.Sprintf("(*%s)", varX)
	}

	varTmpY := varY
	if as.depthY != 0 {
		varTmpY = w.Name(varY)
		w.Printf("var %s %t\n", varTmpY, as.elemY)
	}

	as.assigner.writeAssignCode(w, varX, varTmpY, varErr)

	if as.depthY == 1 {
		w.Printf("%s = &%s\n", varY, varTmpY)
	} else if as.depthY > 1 {
		w.Printf("if %s != nil {\n", varTmpY)
		w.Printf("%s = &%s\n", varY, varTmpY)
		w.Printf("}\n")
	}

	if as.depthX != 0 {
		w.Printf("}\n")
	}
}
