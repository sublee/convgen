//go:build convgen

package main

import (
	"strconv"

	"github.com/sublee/convgen"
)

func ASCII(s string) int {
	return int(s[0])
}

var mod = convgen.Module(
	convgen.ImportFuncErr(strconv.Atoi),
)

type (
	String struct{ X string }
	Int    struct{ X int }
)

var conv = convgen.Struct[String, Int](mod)

func main() {
	// Need strconv.Atoi to convert String.X to Int.X but convgen.Struct
	// injector cannot return an error of strconv.Atoi.
	conv(String{"*"})

	panic("convgen will fail")
}
