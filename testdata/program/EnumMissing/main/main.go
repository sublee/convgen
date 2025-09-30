//go:build convgen

package main

import (
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
	BarA       Bar = "A"
	BarB       Bar = "B"
)

// Foo and Bar cannot be matched without RenameTrimCommonPrefix.
var FooToBar = convgen.Enum[Foo, Bar](nil, BarUnknown)

func main() {
	panic("convgen will fail")
}
