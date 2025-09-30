package assign

import (
	"errors"
	"go/types"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/typeinfo"
)

// keyAssigner assigns a slice, array, or map to a map by converting each key
// and element.
type keyAssigner struct {
	elem         assigner
	key          assigner
	elemX, elemY typeinfo.Type
	keyX, keyY   typeinfo.Type
}

// requiresErr returns true if the underlying element or key assigner returns an
// error.
func (a keyAssigner) requiresErr() bool { return a.elem.requiresErr() || a.key.requiresErr() }

// tryKey tries to create a [keyAssigner] from x to y by converting each key and
// element.
func (fac *factory) tryKey(x, y Object) (*keyAssigner, error) {
	if !y.Type().IsMap() {
		return nil, skip
	}

	var keyX typeinfo.Type
	ok := false
	if x.Type().IsMap() {
		// map[K]V -> map[K2]V2
		ok = true
		keyX = *x.Type().Key
	} else if x.Type().IsSlice() || x.Type().IsArray() {
		if y.Type().Key.IsBasic() && y.Type().Key.Basic.Info()&types.IsInteger != 0 {
			// []T or [N]T -> map[int]T
			ok = true
			keyX = typeinfo.TypeOf(types.Universe.Lookup("int").Type())
		}
	}
	if !ok {
		return nil, skip
	}

	elemX := *x.Type().Elem
	elemY := *y.Type().Elem
	keyY := *y.Type().Key

	var errs error
	elemAs, err := fac.build(typeOnly(fac.inj, elemX), typeOnly(fac.inj, elemY))
	if err != nil {
		errs = errors.Join(errs, err)
	}
	keyAs, err := fac.build(typeOnly(fac.inj, keyX), typeOnly(fac.inj, keyY))
	if err != nil {
		errs = errors.Join(errs, err)
	}
	if errs != nil {
		return nil, errs
	}

	return &keyAssigner{
		elem:  elemAs,
		key:   keyAs,
		elemX: elemX,
		elemY: elemY,
		keyX:  keyX,
		keyY:  keyY,
	}, nil
}

// writeAssignCode writes code that assigns x to y by converting each key and
// element.
func (a keyAssigner) writeAssignCode(w *codefmt.Writer, varX, varY, varErr string) {
	w.Printf("%s = make(map[%t]%t, len(%s))\n", varY, a.keyY, a.elemY, varX)

	varKeyX := w.Name("k")
	varValX := w.Name("v")
	w.Printf("for %s, %s := range %s {\n", varKeyX, varValX, varX)

	varKeyY := w.Name("k")
	w.Printf("var %s %t\n", varKeyY, a.keyY)
	a.key.writeAssignCode(w, varKeyX, varKeyY, varErr)
	if a.requiresErr() {
		w.Printf("if %s != nil {\n", varErr)
		w.Printf("%s = nil\n", varX)
		w.Printf("break\n")
		w.Printf("}\n")
	}

	a.elem.writeAssignCode(w, varValX, varY+"["+varKeyY+"]", varErr)
	if a.requiresErr() {
		w.Printf("if %s != nil {\n", varErr)
		w.Printf("%s = nil\n", varX)
		w.Printf("break\n")
		w.Printf("}\n")
	}

	w.Printf("}\n")
}
