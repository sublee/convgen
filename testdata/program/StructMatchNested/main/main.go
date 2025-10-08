//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	X struct{ S string }
	Y struct{ S string }
)

type (
	XX struct{ ChildX X }
	YY struct{ ChildY Y }
)

type (
	XXPtr struct{ ChildX *X }
)

var MatchNested = convgen.Struct[XX, YY](nil,
	convgen.Match(XX{}.ChildX.S, YY{}.ChildY.S),
	convgen.MatchSkip(XX{}.ChildX, nil),
	convgen.MatchSkip(nil, YY{}.ChildY),
)

var MatchNestedPtr = convgen.Struct[XXPtr, YY](nil,
	convgen.Match(XXPtr{}.ChildX.S, YY{}.ChildY.S),
	convgen.MatchSkip(XXPtr{}.ChildX, nil),
	convgen.MatchSkip(nil, YY{}.ChildY),
)

func main() {
	fmt.Println(MatchNested(XX{X{"hello"}}))
	fmt.Println(MatchNestedPtr(XXPtr{&X{"hello"}}))
	fmt.Println(MatchNestedPtr(XXPtr{nil}))
}
