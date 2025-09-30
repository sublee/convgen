//go:build convgen

package aiwoejfo

import (
	"example.com/UnionNoMethods/other"
	"github.com/sublee/convgen"
)

type (
	X1 struct{}
	Y1 struct{}
)

type (
	X interface{}
	Y interface{}
)

var (
	XtoY = convgen.Union[X, Y](nil)
	XtoZ = convgen.Union[X, other.Z](nil)
)

func main() {
	panic("convgen will fail")
}
