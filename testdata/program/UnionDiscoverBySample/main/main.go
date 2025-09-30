//go:build convgen

package main

import (
	"fmt"

	"example.com/UnionDiscoverBySample/impl"
	"github.com/sublee/convgen"
)

type (
	X interface{ IsX() }
	Y interface{ IsY() }
)

type (
	X1 struct{}
	X2 struct{}
)

func (X1) IsX() {}
func (X2) IsX() {}

var XtoY = convgen.Union[X, Y](nil,
	convgen.DiscoverBySample(nil, impl.Y1{}),
	convgen.RenameTrimCommonPrefix(true, true),
)

func main() {
	// Output: impl.Y1{} impl.Y2{}
	fmt.Printf("%#v\n", XtoY(X1{}))
	fmt.Printf("%#v\n", XtoY(X2{}))
}
