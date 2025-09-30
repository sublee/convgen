//go:build convgen

package testdata

import "github.com/sublee/convgen"

var (
	mod = convgen.Module() // ok
	_   = mod
)

func F1() {
	mod := convgen.Module() // want `cannot assign module to variable inside function`
	_ = mod
}

func F2() {
	func() {
		mod := convgen.Module() // want `cannot assign module to variable inside function`
		_ = mod
	}()
}

type T struct{}

func (T) F3() {
	mod := convgen.Module() // want `cannot assign module to variable inside function`
	_ = mod
}

var F4 = func() {
	mod := convgen.Module() // want `cannot assign module to variable inside function`
	_ = mod
}
