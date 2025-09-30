//go:build convgen

package main

import (
	"fmt"
	"strconv"

	"github.com/sublee/convgen"
)

type X struct {
	Hello int
	World string
	MissX int
}

type Y struct {
	Lorem string
	Ipsum int
	MissY int
}

var XtoY = convgen.StructErr[X, Y](nil,
	convgen.MatchFunc(X{}.Hello, Y{}.Lorem, strconv.Itoa),
	convgen.MatchFuncErr(X{}.World, Y{}.Ipsum, strconv.Atoi),
	convgen.MatchSkip(X{}.MissX, nil),
	convgen.MatchSkip(nil, Y{}.MissY),
)

var XtoYPtr = convgen.StructErr[*X, *Y](nil,
	convgen.MatchFunc(X{}.Hello, Y{}.Lorem, strconv.Itoa),
	convgen.MatchFuncErr(X{}.World, Y{}.Ipsum, strconv.Atoi),
	convgen.MatchSkip(X{}.MissX, nil),
	convgen.MatchSkip(nil, Y{}.MissY),
)

func main() {
	// Output: main.Y{Lorem:"123", Ipsum:456, MissY:0}
	x := X{Hello: 123, World: "456", MissX: 789}
	y, err := XtoY(x)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", y)

	// Output: &main.Y{Lorem:"123", Ipsum:456, MissY:0}
	y2, err2 := XtoYPtr(&x)
	if err2 != nil {
		panic(err2)
	}
	fmt.Printf("%#v\n", y2)

	// Output: converting X.World: strconv.Atoi: parsing "NaN": invalid syntax
	x3 := X{World: "NaN"}
	_, err3 := XtoY(x3)
	fmt.Println(err3)
}
