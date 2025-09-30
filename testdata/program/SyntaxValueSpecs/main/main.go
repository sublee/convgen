//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	X struct{}
	Y struct{}
)

var single = convgen.Struct[X, Y](nil)

var (
	group1 = convgen.Struct[X, Y](nil)
	group2 = convgen.Struct[X, Y](nil)
)

var tuple1, tuple2 = convgen.Struct[X, Y](nil), convgen.Struct[X, Y](nil)

var tuple3, tuple4 = convgen.Struct[X, Y](nil), 42

func main() {
	fmt.Printf("%T\n", single)
	fmt.Printf("%T\n", group1)
	fmt.Printf("%T\n", group2)
	fmt.Printf("%T\n", tuple1)
	fmt.Printf("%T\n", tuple2)
	fmt.Printf("%T\n", tuple3)
	fmt.Printf("%T\n", tuple4)
}
