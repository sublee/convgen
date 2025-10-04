//go:build convgen

package main

import (
	"errors"
	"fmt"

	"github.com/sublee/convgen"
	"github.com/sublee/convgen/pkg/convgenerrors"
)

type Foo int

type Bar string

const (
	BarUnknown Bar = "Unknown"
)

var FooToBar = convgen.EnumErr[Foo, Bar](nil, BarUnknown)

func main() {
	bar, err := FooToBar(Foo(999))

	// Output: Unknown
	fmt.Println(bar)

	// Output: unknown enum member 999: no match found
	fmt.Println(err)

	// Output: true
	fmt.Println(errors.Is(err, convgenerrors.ErrNoMatch))
}
