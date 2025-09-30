//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type Field struct{ Value int }

type Getter struct{ value int }

func (g Getter) Value() int             { return g.value }
func (g Getter) GetValue() int          { return g.value }
func (g Getter) ValueGetter() int       { return g.value }
func (g Getter) DoValueRetrieval() int  { return g.value }
func (g Getter) ValueErr() (int, error) { return g.value, nil }

type Setter struct{ value int }

func (s *Setter) Value(value int)          { s.value = value }
func (s *Setter) SetValue(value int)       { s.value = value }
func (s *Setter) ValueSetter(value int)    { s.value = value }
func (s *Setter) DoValueUpdate(value int)  { s.value = value }
func (s *Setter) ValueErr(value int) error { s.value = value; return nil }

func neg(x int) int { return -x }

var (
	ConvGetter = convgen.Struct[Getter, Field](nil,
		convgen.DiscoverGetters(true, "", ""),
		convgen.MatchSkip(Getter{}.GetValue, nil),
		convgen.MatchSkip(Getter{}.ValueGetter, nil),
		convgen.MatchSkip(Getter{}.DoValueRetrieval, nil),
		convgen.MatchSkip(Getter{}.ValueErr, nil),
	)
	ConvGetterPrefix = convgen.Struct[Getter, Field](nil,
		convgen.DiscoverGetters(true, "Get", ""),
	)
	ConvGetterSuffix = convgen.Struct[Getter, Field](nil,
		convgen.DiscoverGetters(true, "", "Getter"),
	)
	ConvGetterPrefixSuffix = convgen.Struct[Getter, Field](nil,
		convgen.DiscoverGetters(true, "Do", "Retrieval"),
	)
	ConvGetterMatch = convgen.Struct[Getter, Field](nil,
		convgen.Match(convgen.FieldGetter(Getter{}.Value), Field{}.Value),
	)
	ConvGetterMatchFunc = convgen.Struct[Getter, Field](nil,
		convgen.MatchFunc(convgen.FieldGetter(Getter{}.Value), Field{}.Value, neg),
	)
	ConvGetterErr = convgen.StructErr[Getter, Field](nil,
		convgen.DiscoverGetters(true, "", "Err"),
	)
)

var (
	ConvSetter = convgen.Struct[Field, Setter](nil,
		convgen.DiscoverSetters(true, "", ""),
		convgen.MatchSkip(nil, (*Setter)(nil).SetValue),
		convgen.MatchSkip(nil, (*Setter)(nil).ValueSetter),
		convgen.MatchSkip(nil, (*Setter)(nil).DoValueUpdate),
		convgen.MatchSkip(nil, (*Setter)(nil).ValueErr),
	)
	ConvSetterPrefix = convgen.Struct[Field, Setter](nil,
		convgen.DiscoverSetters(true, "Set", ""),
	)
	ConvSetterSuffix = convgen.Struct[Field, Setter](nil,
		convgen.DiscoverSetters(true, "", "Setter"),
	)
	ConvSetterPrefixSuffix = convgen.Struct[Field, Setter](nil,
		convgen.DiscoverSetters(true, "Do", "Update"),
	)
	ConvSetterMatch = convgen.Struct[Field, Setter](nil,
		convgen.Match(Field{}.Value, (*Setter)(nil).Value),
	)
	ConvSetterErr = convgen.StructErr[Field, Setter](nil,
		convgen.DiscoverSetters(true, "", "Err"),
	)
)

func main() {
	g := Getter{value: 42}

	// Output: 42 ... -42 42
	fmt.Println(ConvGetter(g).Value)
	fmt.Println(ConvGetterPrefix(g).Value)
	fmt.Println(ConvGetterSuffix(g).Value)
	fmt.Println(ConvGetterPrefixSuffix(g).Value)
	fmt.Println(ConvGetterMatch(g).Value)
	fmt.Println(ConvGetterMatchFunc(g).Value)
	f, err := ConvGetterErr(g)
	if err != nil {
		panic(err)
	}
	fmt.Println(f.Value)

	// Output: 42 ...
	fmt.Println(ConvSetter(f).value)
	fmt.Println(ConvSetterPrefix(f).value)
	fmt.Println(ConvSetterSuffix(f).value)
	fmt.Println(ConvSetterPrefixSuffix(f).value)
	fmt.Println(ConvSetterMatch(f).value)
	s, err := ConvSetterErr(f)
	if err != nil {
		panic(err)
	}
	fmt.Println(s.value)
}
