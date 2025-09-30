//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	X interface{ isX() }
	Y interface{ isY() }
)

type (
	X0 struct{}
	X1 struct{}
	x2 struct{}
)

type (
	Y0 struct{}
	Y1 struct{}
	y2 struct{}
)

func (X0) isX() {}
func (X1) isX() {}
func (x2) isX() {}
func (Y0) isY() {}
func (Y1) isY() {}
func (y2) isY() {}

var ExportedOnly = convgen.Union[X, Y](nil,
	convgen.RenameToLower(true, true),
	convgen.RenameTrimCommonPrefix(true, true),
)

var UnexportedAlso = convgen.Union[X, Y](nil,
	convgen.RenameToLower(true, true),
	convgen.RenameTrimCommonPrefix(true, true),
	convgen.DiscoverUnexported(true, true),
)

func main() {
	// Output: main.Y0{} main.Y1{} <nil>
	fmt.Printf("%#v\n", ExportedOnly(X0{}))
	fmt.Printf("%#v\n", ExportedOnly(X1{}))
	fmt.Printf("%#v\n", ExportedOnly(x2{}))

	// Output: main.Y0{} main.Y1{} main.y2{}
	fmt.Printf("%#v\n", UnexportedAlso(X0{}))
	fmt.Printf("%#v\n", UnexportedAlso(X1{}))
	fmt.Printf("%#v\n", UnexportedAlso(x2{}))
}
