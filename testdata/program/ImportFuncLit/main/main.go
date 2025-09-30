//go:build convgen

package main

import (
	"fmt"
	"strings"

	"github.com/sublee/convgen"
)

var mod = convgen.Module(
	convgen.ImportFunc(func(n int) int {
		fmt.Println(strings.Repeat("!", n))
		return n
	}),
)

type (
	X struct{ N int }
	Y struct{ N int }
)

var XtoY = convgen.Struct[X, Y](mod)

func main() {
	// Output: !!! 3
	x := X{N: 3}
	y := XtoY(x)
	fmt.Println(y.N)
}
