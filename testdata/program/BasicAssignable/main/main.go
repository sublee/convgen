//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

var conv = convgen.Struct[struct{ X rune }, struct{ X int32 }](nil)

func main() {
	// rune is assignable to int32
	out := conv(struct{ X rune }{X: '*'})

	// Output: 42
	fmt.Println(out.X)
}
