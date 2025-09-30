//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type Human int

const (
	Alice Human = 0
	Bob   Human = 1
)

type Fruit string

const (
	FruitUnknown Fruit = ""
	Apple        Fruit = "A"
	Banana       Fruit = "B"
)

var HumanToFruit = convgen.Enum[Human, Fruit](nil, FruitUnknown,
	// Keep the first letter only.
	convgen.RenameReplaceRegexp("(.).+", "${1}", "(.).+", "${1}"),
)

func main() {
	// Output: A B
	fmt.Println(HumanToFruit(Alice), HumanToFruit(Bob))
}
