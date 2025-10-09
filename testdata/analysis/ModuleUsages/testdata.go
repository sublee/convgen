//go:build convgen

package testdata

import (
	"fmt"

	"github.com/sublee/convgen"
)

var mod = convgen.Module() // ok, very valid

var Mod = convgen.Module() // want `cannot export module "Mod"; removed at code generation`

var _ = convgen.Module() // ok, blank identifier is harmless

var mod2 = mod // want `cannot use module "mod" outside convgen directives; removed at code generation`

var mod3 = &mod // want `cannot use module "mod" outside convgen directives; removed at code generation`

var _ = mod // ok, blank identifier is harmless

func F1() {
	fmt.Println(mod)  // want `cannot use module "mod" outside convgen directives; removed at code generation`
	fmt.Println(mod2) // ok, mod2 is already invalid
	fmt.Println(mod3) // ok, mod3 is already invalid
}

var (
	modSlice = []any{mod} // want `cannot use module "mod" outside convgen directives; removed at code generation`
	_        = modSlice
)

var (
	modArray = [1]any{mod} // want `cannot use module "mod" outside convgen directives; removed at code generation`
	_        = modArray
)

var (
	modMap = map[int]any{0: mod} // want `cannot use module "mod" outside convgen directives; removed at code generation`
	_      = modMap
)

func F2() {
	fmt.Println([]any{mod})          // want `cannot use module "mod" outside convgen directives; removed at code generation`
	fmt.Println([1]any{mod})         // want `cannot use module "mod" outside convgen directives; removed at code generation`
	fmt.Println(map[int]any{0: mod}) // want `cannot use module "mod" outside convgen directives; removed at code generation`
}

func F3() any {
	return mod // want `cannot use module "mod" outside convgen directives; removed at code generation`
}

func F4() {
	Mod := mod // want `cannot use module "mod" outside convgen directives; removed at code generation`
	_ = Mod
}

func F5() {
	mod := convgen.Module() // ok, will be removed
	_ = mod                 // ok, will be removed also
}
