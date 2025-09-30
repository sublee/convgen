//go:build convgen

package main

import (
	"fmt"

	"example.com/EnumMatchUnexported/city"
	"example.com/EnumMatchUnexported/enums"
	"github.com/sublee/convgen"
)

type Alphabet int

const (
	a Alphabet = iota
	B
	C
)

type Name string

const (
	NameUnspecified Name = ""
	Alice           Name = "Alice"
	bob             Name = "Bob"
	Clementine      Name = "Clementine"
)

var ToName = convgen.Enum[Alphabet, Name](nil, NameUnspecified,
	convgen.Match(a, Alice),
	convgen.Match(B, bob),
	convgen.Match(C, Clementine),
)

var ToFruit = convgen.Enum[Alphabet, enums.Fruit](nil, enums.FruitUnspecified,
	convgen.Match(a, enums.Apple),
	convgen.Match(B, enums.Banana),
	convgen.Match(C, enums.Clementine),
)

var ToCity = convgen.Enum[Alphabet, enums.City](nil, enums.CityUnspecified,
	convgen.Match(a, city.Austin),
	convgen.Match(B, city.Boston),
	convgen.Match(C, city.California),
)

func main() {
	// Output: Alice Bob Clementine
	fmt.Println(ToName(a), ToName(B), ToName(C))

	// Output: 1 2 3
	fmt.Println(ToFruit(a), ToFruit(B), ToFruit(C))

	// Output: 1 2 3
	fmt.Println(ToCity(a), ToCity(B), ToCity(C))
}
