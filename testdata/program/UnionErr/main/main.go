//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	X  interface{ x() }
	Y  interface{ y() }
	X0 struct{ S string }
	Y0 struct{ S string }
	x1 struct{ S string }
)

func (X0) x() {}
func (Y0) y() {}
func (x1) x() {}

var XtoY = convgen.UnionErr[X, Y](nil, convgen.RenameTrimPrefix("X", "Y"))

func main() {
	// Output: converting X: unknown type main.x1
	_, err := XtoY(x1{})
	fmt.Println(err)
}
