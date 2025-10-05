//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type Getter struct{ value int }

func (g Getter) Value() int { return g.value }

type Setter struct{ value int }

func (s *Setter) SetValue(value int) { s.value = value }

type (
	X struct{ Child Getter }
	Y struct{ Child *Setter }
)

var WithoutModule = convgen.Struct[X, Y](nil,
	convgen.DiscoverGetters("", ""),
	convgen.DiscoverSetters("Set", ""),
)

var WithModule = convgen.Struct[X, Y](convgen.Module(
	convgen.DiscoverGetters("", ""),
	convgen.DiscoverSetters("Set", ""),
))

func main() {
	// Output: 0
	fmt.Println(WithoutModule(X{Child: Getter{42}}).Child.value)
	// Output: 42
	fmt.Println(WithModule(X{Child: Getter{42}}).Child.value)
}
