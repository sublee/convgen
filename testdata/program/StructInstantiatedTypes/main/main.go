//go:build convgen

package main

import (
	"fmt"
	"strconv"

	"github.com/sublee/convgen"
)

type (
	X[T any] struct{ Value T }
	Y[T any] struct{ Value T }
)

var mod = convgen.Module(convgen.ImportFunc(strconv.Itoa))

var (
	Int8_Int32 = convgen.Struct[X[int8], Y[int32]](mod)
	Int_String = convgen.Struct[X[int], Y[string]](mod)
)

type XX struct {
	Int8_Int32 X[int8]
	Int_String X[int]
}

type YY struct {
	Int8_Int32 Y[int32]
	Int_String Y[string]
}

var XX_YY = convgen.Struct[XX, YY](mod)

func main() {
	// Output: int32 123
	fmt.Printf("%[1]T %[1]v\n", Int8_Int32(X[int8]{Value: 123}).Value)

	// Output: string 123
	fmt.Printf("%[1]T %[1]v\n", Int_String(X[int]{Value: 123}).Value)

	// Output: main.YY{Int8_Int32:main.Y[int32]{Value:123}, Int_String:main.Y[string]{Value:"123"}}
	fmt.Printf("%#v\n", XX_YY(XX{
		Int8_Int32: X[int8]{Value: 123},
		Int_String: X[int]{Value: 123},
	}))
}
