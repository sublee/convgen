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

var Conv1 = convgen.Struct[Getter, Setter](nil,
	convgen.Match(Getter{}.Value, (*Setter)(nil).SetValue),
)

var Conv2 = convgen.Struct[Getter, Setter](nil,
	convgen.Match(Getter{}.Value, (&Setter{}).SetValue),
)

var Conv3 = convgen.Struct[Getter, Setter](nil,
	convgen.Match((&Getter{}).Value, (&Setter{}).SetValue),
)

var Conv4 = convgen.Struct[Getter, Setter](nil,
	convgen.Match((*Getter)(nil).Value, (&Setter{}).SetValue),
)

func main() {
	fmt.Println(Conv1(Getter{1}).value)
	fmt.Println(Conv2(Getter{2}).value)
	fmt.Println(Conv3(Getter{3}).value)
	fmt.Println(Conv4(Getter{4}).value)
}
