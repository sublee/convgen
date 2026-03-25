//go:build convgen

package main

import (
	"fmt"
	"github.com/sublee/convgen"
)

type FooStr struct {
	Value string
}

type BarStr struct {
	Value Generic[string]
}

type FooInt struct {
	Value int
}

type BarInt struct {
	Value Generic[int]
}

type Generic[T any] struct {
	Value T
}

func wrapGeneric[T any](x T) Generic[T] {
	return Generic[T]{Value: x}
}

func wrapGenericErr[T any](x T) (Generic[T], error) {
	return Generic[T]{Value: x}, nil
}

func identity[T any](x T) T {
	return x
}

var mod = convgen.Module(
	convgen.ImportFunc(wrapGeneric[string]),
	convgen.ImportFuncErr(wrapGenericErr[int]),
	convgen.ImportFunc(identity[string]),
)
var IdentityTest = convgen.Struct[FooStr, FooStr](mod)
var WrapTest = convgen.Struct[FooStr, BarStr](mod)
var WrapErrTest = convgen.StructErr[FooInt, BarInt](mod)

func main() {
	foo := IdentityTest(FooStr{"42"})
	fmt.Println(foo.Value)

	bar := WrapTest(FooStr{"42"})
	fmt.Println(bar.Value.Value)

	baz, _ := WrapErrTest(FooInt{42})
	fmt.Println(baz.Value.Value)
}
