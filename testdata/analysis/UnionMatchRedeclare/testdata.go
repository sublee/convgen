//go:build convgen

package testdata

import (
	"github.com/sublee/convgen"
)

type (
	X1 struct{}
	X2 struct{}
	Y1 struct{}
	Y2 struct{}
)

func (X1) isX() {}
func (X2) isX() {}
func (Y1) isY() {}
func (Y2) isY() {}

type (
	X interface{ isX() }
	Y interface{ isY() }
)

var Conv = convgen.Union[X, Y](nil,
	convgen.Match(X1{}, &Y1{}),
	convgen.Match(
		&X1{}, // want `X1 redeclared as pointer type \*X1`
		Y1{},  // want `\*Y1 redeclared as non-pointer type Y1`
	),
	convgen.Match(X2{}, Y2{}),
	convgen.Match(
		&X2{}, // want `X2 redeclared as pointer type \*X2`
		&Y2{}, // want `Y2 redeclared as pointer type \*Y2`
	),
)
