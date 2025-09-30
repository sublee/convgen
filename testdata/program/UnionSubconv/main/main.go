//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	X  interface{ x() }
	Y  interface{ y() }
	X1 struct{ S string }
	Y1 struct{ Str string }
	X2 struct{ N int }
	Y2 struct{ Num int }
)

func (X1) x() {}
func (Y1) y() {}
func (X2) x() {}
func (Y2) y() {}

var mod = convgen.Module(
	convgen.RenameReplaceRegexp("", "", "^(.).*", "${1}"),
)

var XtoY = convgen.Union[X, Y](mod,
	convgen.RenameReset(true, true),
	convgen.RenameTrimPrefix("X", "Y"),
)

func main() {
	// Output: ok 42
	fmt.Println(XtoY(X1{S: "ok"}).(Y1).Str)
	fmt.Println(XtoY(X2{N: 42}).(Y2).Num)
}
