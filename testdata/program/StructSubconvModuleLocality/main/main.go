//go:build convgen

package main

import (
	"fmt"
	"strconv"

	"github.com/sublee/convgen"
)

type (
	X  struct{ Child XX }
	Y  struct{ Child YY }
	XX struct{ N string }
	YY struct{ N int }
)

func Atoi16(s string) (int, error) {
	n, err := strconv.ParseInt(s, 16, 0)
	return int(n), err
}

var (
	mod10 = convgen.Module(convgen.ImportFuncErr(strconv.Atoi))
	mod16 = convgen.Module(convgen.ImportFuncErr(Atoi16))
)

var (
	XtoY10 = convgen.StructErr[X, Y](mod10)
	XtoY16 = convgen.StructErr[X, Y](mod16)
)

func main() {
	// Output: 10 16
	y10, _ := XtoY10(X{Child: XX{N: "10"}})
	y16, _ := XtoY16(X{Child: XX{N: "10"}})
	fmt.Println(y10.Child.N)
	fmt.Println(y16.Child.N)
}
