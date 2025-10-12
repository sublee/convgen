package assign

import (
	"github.com/sublee/convgen/internal/codefmt"
)

// sameAssigner assigns x to y when they are the same type.
type sameAssigner struct{}

func (as sameAssigner) requiresErr() bool { return false }

func (fac *factory) trySame(x, y Object) (*sameAssigner, error) {
	if !x.Type().Identical(y.Type()) {
		return nil, skip
	}

	return &sameAssigner{}, nil
}

// writeAssignCode writes code that assigns x to y by converting each element.
func (as sameAssigner) writeAssignCode(w *codefmt.Writer, varX, varY, varErr string) {
	w.Printf("%s = %s\n", varY, varX)
}
