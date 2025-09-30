//go:build convgen

package main

import (
	"strconv"

	"github.com/sublee/convgen"
)

type (
	X struct{}
	Y struct{}
)

var mod = convgen.Module(
	convgen.ImportFunc(strconv.Itoa),
	convgen.ImportFunc(strconv.Itoa),
	convgen.ImportFunc(xToY),
)

func xToY(X) Y { return Y{} }

var (
	XtoY  = convgen.Struct[X, Y](mod)
	YtoX  = convgen.Struct[Y, X](mod)
	YtoX2 = convgen.Struct[Y, X](mod)
)

func main() {
	panic("convgen will fail")
}
