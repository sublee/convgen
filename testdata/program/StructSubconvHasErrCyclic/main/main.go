//go:build convgen

package main

import (
	fmt "fmt"
	"strconv"

	"github.com/sublee/convgen"
)

type (
	X struct{ Child1 XX }
	Y struct{ Child1 YY }

	XX struct {
		Child2 XXX
		Atoi   string
	}
	YY struct {
		Child2 YYY
		Atoi   int
	}

	XXX struct{ Child3 *XX }
	YYY struct{ Child3 *YY }
)

var (
	mod  = convgen.Module(convgen.ImportFuncErr(strconv.Atoi))
	XtoY = convgen.StructErr[X, *Y](mod)
)

func main() {
	_, err := XtoY(X{Child1: XX{Child2: XXX{Child3: &XX{Child2: XXX{nil}, Atoi: "ERROR"}}, Atoi: "42"}})
	// Output: converting X.Child1.Child2.Child3.Atoi: strconv.Atoi: parsing "ERROR": invalid syntax
	fmt.Println(err)
}
