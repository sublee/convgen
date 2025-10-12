//go:build convgen

package main

import (
	"fmt"
	"time"

	"github.com/sublee/convgen"
)

type (
	X struct {
		S string
		T time.Time
	}
	Y struct {
		S *string
		T *time.Time
	}
)

var (
	XtoY = convgen.Struct[X, Y](nil)
	YtoX = convgen.Struct[Y, X](nil)
)

func main() {
	// Output: "hello"
	// Output: 2025-01-02 03:04:05 +0000 UTC
	y := XtoY(X{S: "hello", T: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)})
	fmt.Printf("%q\n", *y.S)
	fmt.Printf("%v\n", y.T)

	// Output: ""
	// Output: 0001-01-01 00:00:00 +0000 UTC
	y0 := XtoY(X{S: "", T: time.Time{}})
	fmt.Printf("%q\n", *y0.S)
	fmt.Printf("%v\n", y0.T)

	// Output: main.X{S:"hello", T:time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)}
	s := "hello"
	t := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	fmt.Printf("%#v\n", YtoX(Y{S: &s, T: &t}))

	// Output: main.X{S:"", T:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
	fmt.Printf("%#v\n", YtoX(Y{S: nil, T: nil}))
}
