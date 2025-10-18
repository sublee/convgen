//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type X struct {
	Slice []int
	Map   map[int]int
}

type Y X

var XtoY = convgen.Struct[X, Y](nil)

func main() {
	// Output: true true
	xNil := X{}
	yNil := XtoY(xNil)
	fmt.Println(yNil.Slice == nil, yNil.Map == nil)

	// Output: true true
	xLen0 := X{Slice: []int{}, Map: map[int]int{}}
	yLen0 := XtoY(xLen0)
	fmt.Println(yLen0.Slice == nil, yLen0.Map == nil)
}
