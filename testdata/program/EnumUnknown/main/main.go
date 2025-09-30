//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type Foo int

type Bar string

const (
	BarUnknown Bar = "Unknown"
)

var FooToBar = convgen.Enum[Foo, Bar](nil, BarUnknown)

func main() {
	// Output: Unknown
	fmt.Println(FooToBar(Foo(999)))
}
