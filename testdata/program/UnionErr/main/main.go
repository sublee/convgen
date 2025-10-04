//go:build convgen

package main

import (
	"errors"
	"fmt"

	"github.com/sublee/convgen"
	"github.com/sublee/convgen/pkg/convgenerrors"
)

type (
	X  interface{ x() }
	Y  interface{ y() }
	X0 struct{ S string }
	Y0 struct{ S string }
	x1 struct{ S string }
)

func (X0) x() {}
func (Y0) y() {}
func (x1) x() {}

var XtoY = convgen.UnionErr[X, Y](nil, convgen.RenameTrimPrefix("X", "Y"))

func main() {
	// Output: converting X: unknown union impl main.x1: no match found
	_, err := XtoY(x1{})
	fmt.Println(err)

	// Output: true
	fmt.Println(errors.Is(err, convgenerrors.ErrNoMatch))
}
