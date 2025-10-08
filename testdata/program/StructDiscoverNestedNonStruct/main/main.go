//go:build convgen

package main

import (
	"github.com/sublee/convgen"
)

type (
	X struct {
		A int
		B *int
	}
	Y struct{}
)

var XtoY = convgen.Struct[X, Y](nil,
	convgen.DiscoverNested(X{}.A, nil),
	convgen.DiscoverNested(X{}.B, nil),
)

func main() {
	panic("convgen will fail")
}
