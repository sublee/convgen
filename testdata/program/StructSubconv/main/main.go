//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	XX  struct{ Child XXX }
	YY  struct{ Child YYY }
	XXX struct{ String string }
	YYY struct{ String string }
)

type X struct {
	Value      XX
	ValueToPtr XX
	PtrToValue *XX
	Ptr        *XX
}

type Y struct {
	Value      YY
	ValueToPtr *YY
	PtrToValue YY
	Ptr        *YY
}

var XtoY = convgen.Struct[X, Y](nil)

func main() {
	// Output: a b c d
	x := X{
		Value:      XX{XXX{String: "a"}},
		ValueToPtr: XX{XXX{String: "b"}},
		PtrToValue: &XX{XXX{String: "c"}},
		Ptr:        &XX{XXX{String: "d"}},
	}
	y := XtoY(x)
	fmt.Println(y.Value.Child.String)
	fmt.Println(y.ValueToPtr.Child.String)
	fmt.Println(y.PtrToValue.Child.String)
	fmt.Println(y.Ptr.Child.String)

	// Output: true true
	x2 := X{PtrToValue: nil, Ptr: nil}
	y2 := XtoY(x2)
	fmt.Println(y2.PtrToValue.Child.String == "")
	fmt.Println(y2.Ptr == nil)
}
