//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type Foo int

const (
	FooA Foo = 0
	FooB Foo = 1
)

type Bar string

const (
	UnknownBar Bar = ""
	ABar       Bar = "A"
	BBar       Bar = "B"
)

var FooToBar = convgen.Enum[Foo, Bar](nil, UnknownBar,
	convgen.RenameTrimCommonWordPrefix(true, false),
	convgen.RenameTrimCommonWordSuffix(false, true),
)

func main() {
	// Output: A B
	fmt.Println(FooToBar(FooA), FooToBar(FooB))
}
