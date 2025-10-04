//go:build convgen

package main

import (
	"github.com/sublee/convgen"
)

type (
	X interface{ IsX() }
	Y interface{ IsY() }
)

var XtoY = convgen.Union[X, Y](nil,
	convgen.DiscoverBySample(nil, nil),
)

func main() {
	panic("convgen will fail")
}
