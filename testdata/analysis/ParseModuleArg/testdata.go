//go:build convgen

package testdata

import (
	"github.com/sublee/convgen"
)

var C1 = convgen.Struct[struct{}, struct{ a int }](
	nil, // ok
)

var C2 = convgen.Struct[struct{}, struct{ b int }](
	convgen.Module(), // ok
)

var mod2 = convgen.Module()

var C3 = convgen.Struct[struct{}, struct{ c int }](
	mod2, // ok
)

var C4 = convgen.Struct[struct{}, struct{ d int }](
	(mod2), // ok
)

func asis[T any](T) T { return *new(T) }

var C5 = convgen.Struct[struct{}, struct{ e int }](
	asis(convgen.Module()), // want `module must be convgen.Module\(\) or package-level variable`
)

var C6 = convgen.Struct[struct{}, struct{ f int }](
	(asis(convgen.Module())), // want `module must be convgen.Module\(\) or package-level variable`
)
