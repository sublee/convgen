//go:build convgen

package main

import (
	fmt "fmt"

	"github.com/sublee/convgen"
)

type (
	X  struct{ Child *XX }
	Y  struct{ Child *YY }
	XX struct{ Parent *X }
	YY struct{ Parent *Y }
)

var (
	XtoY       = convgen.Struct[X, Y](nil)
	PtrXtoY    = convgen.Struct[*X, Y](nil)
	XtoPtrY    = convgen.Struct[X, *Y](nil)
	PtrXtoPtrY = convgen.Struct[*X, *Y](nil)
)

func main() {
	// XtoY

	x1 := X{&XX{}}
	y1 := XtoY(x1)
	// Output: true true
	fmt.Println(y1.Child != nil)
	fmt.Println(y1.Child.Parent == nil)

	// Output: true true true
	x2 := X{&XX{&X{}}}
	y2 := XtoY(x2)
	fmt.Println(y2.Child != nil)
	fmt.Println(y2.Child.Parent != nil)
	fmt.Println(y2.Child.Parent.Child == nil)

	// Output: true true true true
	x3 := X{&XX{&X{&XX{}}}}
	y3 := XtoY(x3)
	fmt.Println(y3.Child != nil)
	fmt.Println(y3.Child.Parent != nil)
	fmt.Println(y3.Child.Parent.Child != nil)
	fmt.Println(y3.Child.Parent.Child.Parent == nil)

	// PtrXtoY

	x4 := &X{&XX{}}
	y4 := PtrXtoY(x4)
	// Output: true true
	fmt.Println(y4.Child != nil)
	fmt.Println(y4.Child.Parent == nil)

	// Output: true true true
	x5 := &X{&XX{&X{}}}
	y5 := PtrXtoY(x5)
	fmt.Println(y5.Child != nil)
	fmt.Println(y5.Child.Parent != nil)
	fmt.Println(y5.Child.Parent.Child == nil)

	// Output: true true true true
	x6 := &X{&XX{&X{&XX{}}}}
	y6 := PtrXtoY(x6)
	fmt.Println(y6.Child != nil)
	fmt.Println(y6.Child.Parent != nil)
	fmt.Println(y6.Child.Parent.Child != nil)
	fmt.Println(y6.Child.Parent.Child.Parent == nil)

	// XtoPtrY

	x7 := X{&XX{}}
	y7 := XtoPtrY(x7)
	// Output: true true
	fmt.Println(y7.Child != nil)
	fmt.Println(y7.Child.Parent == nil)

	// Output: true true true
	x8 := X{&XX{&X{}}}
	y8 := XtoPtrY(x8)
	fmt.Println(y8.Child != nil)
	fmt.Println(y8.Child.Parent != nil)
	fmt.Println(y8.Child.Parent.Child == nil)

	// Output: true true true true
	x9 := X{&XX{&X{&XX{}}}}
	y9 := XtoPtrY(x9)
	fmt.Println(y9.Child != nil)
	fmt.Println(y9.Child.Parent != nil)
	fmt.Println(y9.Child.Parent.Child != nil)
	fmt.Println(y9.Child.Parent.Child.Parent == nil)

	// PtrXtoPtrY

	x10 := &X{&XX{}}
	y10 := PtrXtoPtrY(x10)
	// Output: true true
	fmt.Println(y10.Child != nil)
	fmt.Println(y10.Child.Parent == nil)

	// Output: true true true
	x11 := &X{&XX{&X{}}}
	y11 := PtrXtoPtrY(x11)
	fmt.Println(y11.Child != nil)
	fmt.Println(y11.Child.Parent != nil)
	fmt.Println(y11.Child.Parent.Child == nil)

	// Output: true true true true
	x12 := &X{&XX{&X{&XX{}}}}
	y12 := PtrXtoPtrY(x12)
	fmt.Println(y12.Child != nil)
	fmt.Println(y12.Child.Parent != nil)
	fmt.Println(y12.Child.Parent.Child != nil)
	fmt.Println(y12.Child.Parent.Child.Parent == nil)
}
