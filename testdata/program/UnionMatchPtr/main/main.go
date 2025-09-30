//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	X1 struct{}
	X2 struct{}
	Y1 struct{}
	Y2 struct{}
)

func (X1) isX()  {}
func (X2) isX()  {}
func (*Y1) isY() {}
func (Y2) isY()  {}

type (
	X interface{ isX() }
	Y interface{ isY() }
)

var Conv1 = convgen.Union[X, Y](nil,
	convgen.RenameTrimPrefix("X", "Y"),
)

var Conv2 = convgen.Union[X, Y](nil,
	convgen.RenameTrimPrefix("X", "Y"),
	convgen.Match(&X1{}, &Y1{}),
)

func main() {
	// Output: &main.Y1{}
	fmt.Printf("%#v\n", Conv1(X1{}))

	// Output: <nil>
	fmt.Printf("%#v\n", Conv1(&X1{}))

	// Output: <nil>
	fmt.Printf("%#v\n", Conv2(X1{}))

	// Output: &main.Y1{}
	fmt.Printf("%#v\n", Conv2(&X1{}))
}
