//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type X struct {
	Array      [2]int
	Slice      []int
	Map        map[int]int
	ArrayToMap [3]int
	SliceToMap []int
}

type Y struct {
	Array      [2]int
	Slice      []int
	Map        map[int]int
	ArrayToMap map[int]int
	SliceToMap map[int]int
}

var XtoY = convgen.Struct[X, Y](nil)

func main() {
	x := X{
		Array:      [2]int{1, 2},
		Slice:      []int{3, 4, 5},
		Map:        map[int]int{6: 7, 8: 9},
		ArrayToMap: [3]int{10, 11, 12},
		SliceToMap: []int{13, 14},
	}
	y := XtoY(x)

	// Output: {String:hello Int:42 Array:[97 98 99] Slice:[100 101 102] Map:map[103:14]}0
	fmt.Printf("%+v\n", y)
}
