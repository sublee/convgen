//go:build convgen

package main

import (
	"fmt"
	"strconv"

	"github.com/sublee/convgen"
)

// A  B
// |  |
// |  C
// | /|
// D  Atoi
// |
// Itoa
type (
	Ax struct{ Next Dx }
	Bx struct{ Next Cx }
	Cx struct {
		Next Dx
		Atoi string
	}
	Dx struct{ Itoa int }

	Ay struct{ Next Dy }
	By struct{ Next Cy }
	Cy struct {
		Next Dy
		Atoi int
	}
	Dy struct{ Itoa string }
)

var mod = convgen.Module(
	convgen.ImportFunc(strconv.Itoa),
	convgen.ImportFuncErr(strconv.Atoi),
)

var (
	ConvA = convgen.Struct[Ax, Ay](mod)
	ConvB = convgen.StructErr[Bx, By](mod)
)

func main() {
	// Output: true
	a := ConvA(Ax{Next: Dx{Itoa: 42}})
	fmt.Println(a.Next.Itoa == "42")

	// Output: true true
	b, err := ConvB(Bx{Next: Cx{Atoi: "42", Next: Dx{Itoa: 42}}})
	if err != nil {
		panic(err)
	}
	fmt.Println(b.Next.Atoi == 42)
	fmt.Println(b.Next.Next.Itoa == "42")

	// Output: converting Next: converting Atoi: strconv.Atoi: parsing "Forty-two": invalid syntax
	_, err2 := ConvB(Bx{Next: Cx{Atoi: "Forty-two", Next: Dx{Itoa: 42}}})
	fmt.Println(err2)
}
