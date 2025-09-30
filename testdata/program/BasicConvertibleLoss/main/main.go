//go:build convgen

package main

import (
	"github.com/sublee/convgen"
)

var conv = convgen.Struct[struct{ X int32 }, struct{ X int8 }](nil)

func main() {
	// int32 is not convertible to int8 without loss of information
	conv(struct{ X int32 }{X: 42})

	panic("convgen will fail")
}
