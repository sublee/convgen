package assign

import (
	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/typeinfo"
)

// errWrapAssigner wraps a conversion error with user-defined functions.
type errWrapAssigner struct {
	typeinfo.Func
	next *errWrapAssigner
}

// newErrWrap creates a chain of [errWrapAssigner] from the
// [convgen.ImportErrWrap] configurations.
func (fac *factory) newErrWrap() *errWrapAssigner {
	if !fac.allowsErr {
		// If this factory does not allow returning error, errWrap is useless.
		// So, we skip creating it.
		return nil
	}

	as := &errWrapAssigner{}
	for _, fn := range fac.cfg.ErrWraps {
		as = &errWrapAssigner{
			Func: fn,
			next: as,
		}
	}
	return as
}

// requiresErr always returns true.
func (as errWrapAssigner) requiresErr() bool { return true }

// writeAssignCode writes code that wraps the error in varX and assigns it to
// varY. If varX is the same as varY, it only wraps the error in place.
func (as errWrapAssigner) writeAssignCode(w *codefmt.Writer, varX, _, varErr string) {
	if as.Func == nil {
		if varX != varErr {
			w.Printf("%s = %s\n", varErr, varX)
		}
		return
	}

	if as.next != nil {
		as.next.writeWrapCode(w, varX)
	}

	w.Printf("%s = ", varErr)
	if as.Name() != "" {
		w.Printf("%o", as.Func)
	} else if as.FuncLit() != nil {
		w.Printf("%c", as.Func.FuncLit())
	}
	w.Printf("(%s)\n", varX)
}

// writeWrapCode is a helper that calls [errWrapAssigner.writeAssignCode] with
// varX and varErr being the same.
func (as errWrapAssigner) writeWrapCode(w *codefmt.Writer, varErr string) {
	as.writeAssignCode(w, varErr, "", varErr)
}
