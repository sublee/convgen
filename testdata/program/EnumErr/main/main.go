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

var FooToBar = convgen.EnumErr[Foo, Bar](nil, BarUnknown)

func main() {
	bar, err := FooToBar(Foo(999))

	// Output: Unknown
	fmt.Println(bar)

	// Output: cannot convert Foo value 999 to Bar
	fmt.Println(err)
}
