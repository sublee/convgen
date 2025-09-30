//go:build convgen

package testdata

import "github.com/sublee/convgen"

func F1() {
	_ = convgen.Module() // want `cannot assign module to variable inside function`
}
