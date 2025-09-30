//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type X struct {
	String string
	Int    int
	Array  [3]rune
	Slice  []rune
	Map    map[rune]rune
}

type Y struct {
	String string
	Int    int
	Array  [3]rune
	Slice  []rune
	Map    map[rune]rune
}

var XtoY = convgen.Struct[X, Y](nil)

func main() {
	x := X{
		String: "hello",
		Int:    42,
		Array:  [3]rune{'a', 'b', 'c'},
		Slice:  []rune{'d', 'e', 'f'},
		Map:    map[rune]rune{'g': 'h'},
	}
	y := XtoY(x)

	// Output: {String:hello Int:42 Array:[97 98 99] Slice:[100 101 102] Map:map[103:104]}
	fmt.Printf("%+v\n", y)
}
