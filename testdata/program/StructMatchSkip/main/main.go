//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type X struct {
	Foo int
	Baz int
}

type Y struct {
	Bar int
	Baz int
}

var SkipAll = convgen.Struct[X, Y](nil,
	convgen.MatchSkip(X{}.Foo, nil),
	convgen.MatchSkip(nil, Y{}.Bar),
	convgen.MatchSkip(X{}.Baz, Y{}.Baz),
)

var SkipMissing = convgen.Struct[X, Y](nil,
	convgen.MatchSkip(X{}.Foo, nil),
	convgen.MatchSkip(nil, Y{}.Bar),
)

var SkipMissingPtr = convgen.Struct[*X, *Y](nil,
	convgen.MatchSkip(X{}.Foo, nil),
	convgen.MatchSkip(nil, Y{}.Bar),
)

func main() {
	// Output: main.Y{Bar:0, Baz:0}
	fmt.Printf("%#v\n", SkipAll(X{Foo: 1, Baz: 2}))

	// Output: main.Y{Bar:0, Baz:4}
	fmt.Printf("%#v\n", SkipMissing(X{Foo: 3, Baz: 4}))

	// Output: &main.Y{Bar:0, Baz:6}
	fmt.Printf("%#v\n", SkipMissingPtr(&X{Foo: 5, Baz: 6}))
}
