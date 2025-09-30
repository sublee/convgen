//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	StructX[T any] struct{}
	StructY[T any] struct{}
	UnionX[T any]  interface{ isX() }
	UnionY[T any]  interface{ isY() }
)

func (StructX[T]) isX() {}
func (StructY[T]) isY() {}

var (
	StructGeneric = convgen.Struct[StructX[string], StructY[string]](nil)
	UnionGeneric  = convgen.Union[UnionX[string], UnionY[string]](nil,
		convgen.RenameTrimSuffix("X", "Y"),
	)
)

func main() {
	fmt.Printf("%#v\n", StructGeneric(StructX[string]{}))
	fmt.Printf("%#v\n", UnionGeneric((UnionX[string])(nil)))
}
