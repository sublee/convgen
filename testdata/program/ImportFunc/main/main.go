//go:build convgen

package main

import (
	"fmt"
	"strconv"

	"github.com/sublee/convgen"
)

var mod = convgen.Module(
	convgen.ImportFunc(strconv.Itoa),
	convgen.ImportFuncErr(strconv.Atoi),
)

type (
	Int    struct{ X int }
	String struct{ X string }
)

var (
	Itoa = convgen.Struct[Int, String](mod)
	Atoi = convgen.StructErr[String, Int](mod)
)

func main() {
	// Output: 42
	s := Itoa(Int{42})
	fmt.Println(s.X)

	// Output: converting String.X: strconv.Atoi: parsing "*": invalid syntax
	_, err := Atoi(String{"*"})
	fmt.Println(err)
}
