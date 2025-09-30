//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	X  struct{ Child XX }
	Y  struct{ Underlying YY }
	XX struct{ I int }
	YY struct{ I int }
)

var XtoY = convgen.Struct[X, Y](nil,
	convgen.Match(X{}.Child, Y{}.Underlying),
)

func main() {
	x := X{XX{42}}
	y := XtoY(x)

	// Output: 42
	fmt.Println(y.Underlying.I)
}
