//go:build convgen

package testdata

import (
	"github.com/sublee/convgen"
)

type (
	X   struct{ S string }
	Y   struct{ S string }
	XX  struct{ Grandchild X }
	YY  struct{ Grandchild Y }
	XXX struct{ Child XX }
	YYY struct{ Child YY }
)

var C1 = convgen.Struct[X, YY](nil,
	convgen.DiscoverNested(nil, YY{}.Grandchild), // ok
)

var C2 = convgen.Struct[XX, Y](nil,
	convgen.DiscoverNested(XX{}.Grandchild, nil), // ok
)

var C6 = convgen.Struct[XX, Y](nil,
	convgen.DiscoverNested((XX{}.Grandchild), (nil)), // ok
)

var C3 = convgen.Struct[XXX, Y](nil,
	convgen.MatchSkip(XXX{}.Child, nil),                 // ok
	convgen.DiscoverNested(XXX{}.Child.Grandchild, nil), // ok
)

var C4 = convgen.Struct[XX, YY](nil,
	convgen.MatchSkip(XX{}.Grandchild.S, YY{}.Grandchild.S), // ok
)

var C5 = convgen.Struct[XXX, Y](nil,
	convgen.MatchSkip(XXX{}.Child, nil),
	convgen.Match(XXX{}.Child.Grandchild.S, Y{}.S),
)
