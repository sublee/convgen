//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	X struct {
		A  int
		C1 *XX
	}
	X0 struct {
		C1 *XX
	}
	XX struct {
		B  int
		C2 **XXX
	}
	XXX struct {
		C int
	}

	Y struct {
		A, B, C int
	}
)

var XtoY = convgen.Struct[X, Y](nil,
	convgen.DiscoverNested(X{}.C1, nil),
	convgen.DiscoverNested(X{}.C1.C2, nil),
)

var YtoX = convgen.Struct[Y, X](nil,
	convgen.DiscoverNested(nil, X{}.C1),
	convgen.DiscoverNested(nil, X{}.C1.C2),
)

var XXtoY = convgen.Struct[XX, Y](nil,
	convgen.MatchSkip(nil, Y{}.A),
	convgen.DiscoverNested(XX{}.C2, nil),
)

var YtoXX = convgen.Struct[Y, XX](nil,
	convgen.MatchSkip(Y{}.A, nil),
	convgen.DiscoverNested(nil, XX{}.C2),
)

var X0toY = convgen.Struct[X0, Y](nil,
	convgen.MatchSkip(nil, Y{}.A),
	convgen.DiscoverNested(X0{}.C1, nil),
	convgen.DiscoverNested((*(X0{}.C1)).C2, nil),
)

var YtoX0 = convgen.Struct[Y, X0](nil,
	convgen.MatchSkip(Y{}.A, nil),
	convgen.DiscoverNested(nil, X0{}.C1),
	convgen.DiscoverNested(nil, (*(X0{}.C1)).C2),
)

func main() {
	xxx := &XXX{C: 3}
	x := X{A: 1, C1: &XX{B: 2, C2: &xxx}}
	y := XtoY(x)
	fmt.Println(y)                                             // Output: {1 2 3}
	fmt.Println(YtoX(y).A, YtoX(y).C1.B, (*(YtoX(y).C1.C2)).C) // Output: 1 2 3
	fmt.Println(XXtoY(*x.C1))                                  // Output: {0 2 3}
	fmt.Println(YtoXX(y).B, (*(YtoXX(y).C2)).C)                // Output: 2 3
	fmt.Println(X0toY(X0{C1: &XX{B: 2, C2: &xxx}}))            // Output: {0 2 3}
	fmt.Println(YtoX0(y).C1.B, (*(YtoX0(y).C1.C2)).C)          // Output: 2 3
}
