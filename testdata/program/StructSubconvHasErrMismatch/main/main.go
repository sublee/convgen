//go:build convgen

package main

import (
	"strconv"

	"github.com/sublee/convgen"
)

type (
	X1 struct{ Child XX }
	Y1 struct{ Child YY }

	X2 struct{ Child XX }
	Y2 struct{ Child YY }

	XX struct{ S string }
	YY struct{ S int }
)

var mod = convgen.Module(convgen.ImportFuncErr(strconv.Atoi))

// Both implicitly requires an XX->YY subconverter. The subconverter has an
// error because of Atoi. But XtoY2 does not return an error, so Convgen should
// fail.
var (
	XtoY1 = convgen.StructErr[X1, Y1](mod)
	XtoY2 = convgen.Struct[X2, Y2](mod)
)

func main() {
	panic("convgen will fail")
}
