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
	X1 struct{ I int }
	X2 struct{ B bool }
	Y0 struct{ S string }
	Y1 struct{ I int }
	Y2 struct{ B bool }
)

func (X0) x() {}
func (X1) x() {}
func (X2) x() {}
func (Y0) y() {}
func (Y1) y() {}
func (Y2) y() {}

var XtoY = convgen.Union[X, Y](nil, convgen.RenameTrimPrefix("X", "Y"))

func main() {
	// Output: hello
	y0 := XtoY(X0{"hello"})
	fmt.Println(y0.(Y0).S)

	// Output: 42
	y1 := XtoY(X1{42})
	fmt.Println(y1.(Y1).I)

	// Output: true
	y2 := XtoY(X2{true})
	fmt.Println(y2.(Y2).B)

	// Output: <nil>
	y3 := XtoY(nil)
	fmt.Println(y3)
}
