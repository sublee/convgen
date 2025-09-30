//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

var conv = convgen.Struct[struct{ X int8 }, struct{ X int32 }](nil)

func main() {
	// int8 is convertible to int32 without loss of information
	out := conv(struct{ X int8 }{X: 42})

	// Output: 42
	fmt.Println(out.X)
}
