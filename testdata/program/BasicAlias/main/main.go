//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type myInt = int

var conv = convgen.Struct[struct{ X myInt }, struct{ X int }](nil)

func main() {
	// myInt is assignable to int
	out := conv(struct{ X myInt }{X: myInt(42)})

	// Output: 42
	fmt.Println(out.X)
}
