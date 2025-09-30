//go:build convgen

package main

import (
	"fmt"

	"example.com/EnumExact/bar"
	"example.com/EnumExact/foo"
	"github.com/sublee/convgen"
)

var FooToBar = convgen.Enum[foo.Foo, bar.Bar](nil, bar.BarUnknown)

func main() {
	// Output: A B C
	fmt.Println(FooToBar(foo.A), FooToBar(foo.B), FooToBar(foo.C))
}
