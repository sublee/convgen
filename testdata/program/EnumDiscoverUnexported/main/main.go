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
	BarUnknown Bar = "Unknown"
	barA       Bar = "A"
	barB       Bar = "B"
)

var ExportedOnly = convgen.Enum[Foo, Bar](nil, BarUnknown,
	convgen.MatchSkip(FooA, nil),
	convgen.MatchSkip(FooB, nil),
)

var UnexportedToo = convgen.Enum[Foo, Bar](nil, BarUnknown,
	convgen.RenameTrimPrefix("Foo", "bar"),
	convgen.DiscoverUnexported(false, true),
)

func main() {
	// Output: Unknown Unknown
	fmt.Println(ExportedOnly(FooA), ExportedOnly(FooB))

	// Output: A B
	fmt.Println(UnexportedToo(FooA), UnexportedToo(FooB))
}
