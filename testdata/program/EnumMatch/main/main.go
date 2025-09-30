//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type Alphabet int

const (
	A Alphabet = iota
	B
)

type Name string

const (
	NameUnspecified Name = ""
	Alice           Name = "Alice"
	Bob             Name = "Bob"
)

var ToName = convgen.Enum[Alphabet, Name](nil, NameUnspecified,
	convgen.Match(A, Alice),
	convgen.Match(B, Bob),
)

func main() {
	// Output: Alice Bob
	fmt.Println(ToName(A), ToName(B))
}
