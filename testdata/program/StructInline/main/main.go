//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

var (
	ConvInline    = convgen.Struct[struct{ I int8 }, struct{ I int32 }](nil)
	ConvInlinePtr = convgen.Struct[*struct{ I int8 }, *struct{ I int32 }](nil)
)

func main() {
	// Output: 42
	fmt.Println(ConvInline(struct{ I int8 }{I: 42}).I)

	// Output: 123
	x := struct{ I int8 }{I: 123}
	y := ConvInlinePtr(&x)
	fmt.Println((*y).I)
}
