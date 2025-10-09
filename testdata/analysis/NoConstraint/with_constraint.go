//go:build convgen

package testdata

import "github.com/sublee/convgen"

func F1() {
	_ = convgen.RenameReset(true, true) // want `cannot assign RenameReset to variable`
}
