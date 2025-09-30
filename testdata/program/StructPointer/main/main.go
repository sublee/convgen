//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	X struct{ String string }
	Y struct{ String string }
)

var mod = convgen.Module()

var (
	XtoPtrY    = convgen.Struct[X, *Y](mod)
	PtrXtoY    = convgen.Struct[*X, Y](mod)
	PtrXtoPtrY = convgen.Struct[*X, *Y](mod)
	PtrPtrXtoY = convgen.Struct[**X, Y](mod)
	XtoPtrPtrY = convgen.Struct[X, **Y](mod)
)

func main() {
	fmt.Println((*XtoPtrY(X{"one"})).String)
	fmt.Println(PtrXtoY(&X{"two"}).String)
	fmt.Println((*PtrXtoPtrY(&X{"three"})).String)

	x := &X{"four"}
	fmt.Println(PtrPtrXtoY(&x).String)

	fmt.Println((*(*XtoPtrPtrY(X{"five"}))).String)
}
