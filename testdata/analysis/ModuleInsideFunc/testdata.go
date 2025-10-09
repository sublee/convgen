//go:build convgen

package testdata

import "github.com/sublee/convgen"

var (
	mod = convgen.Module() // ok
	_   = mod
)

func F1() {
	mod := convgen.Module() // ok
	_ = mod
}

func F2() {
	func() {
		mod := convgen.Module() // ok
		_ = mod
	}()
}

type T struct{}

func (T) F3() {
	mod := convgen.Module() // ok
	_ = mod
}

var F4 = func() {
	mod := convgen.Module() // ok
	_ = mod
}

// Previously, assigning module to variable inside function was disallowed. Now
// it's allowed, unless the variable is exported or used at runtime.
