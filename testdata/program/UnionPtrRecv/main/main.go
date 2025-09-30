//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	X1 struct{}
	X2 struct{}
	Y1 struct{}
	Y2 struct{}
)

func (*X1) isX() {}
func (X2) isX()  {}
func (Y1) isY()  {}
func (*Y2) isY() {}

type (
	X interface{ isX() }
	Y interface{ isY() }
)

var XtoY = convgen.Union[X, Y](nil, convgen.RenameTrimCommonPrefix(true, true))

func main() {
	fmt.Printf("%#v\n", XtoY(&X1{}))
	fmt.Printf("%#v\n", XtoY(X2{}))
}
