//go:build convgen

package main

import (
	"fmt"
	"strconv"

	"github.com/sublee/convgen"
)

func wrapErr(err error) error {
	return fmt.Errorf("A: %w", err)
}

var mod = convgen.Module(
	convgen.ImportErrWrap(wrapErr),
	convgen.ImportErrWrap(func(err error) error { return fmt.Errorf("B: %w", err) }),
	convgen.ImportFuncErr(strconv.Atoi),
)

type (
	X struct{ Atoi string }
	Y struct{ Atoi int }
)

var conv = convgen.StructErr[X, Y](mod)

func main() {
	// Output: B: A: converting X.Atoi: strconv.Atoi: parsing "abc": invalid syntax
	_, err := conv(X{"abc"})
	fmt.Println(err.Error())
}
