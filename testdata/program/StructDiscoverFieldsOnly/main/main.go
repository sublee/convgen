//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

type X struct {
	name string
	Age  int
}

func (x X) GetName() string { return reverse(x.name) }

type Y struct {
	name string
	Age  int
}

func (y *Y) SetName(name string) { y.name = reverse(name) }

var GetterSetter = convgen.Struct[X, Y](convgen.Module(
	convgen.DiscoverGetters("Get", ""),
	convgen.DiscoverSetters("Set", ""),
))

var GetterOnly = convgen.Struct[X, Y](convgen.Module(
	convgen.DiscoverGetters("Get", ""),
	convgen.DiscoverSetters("Set", ""),
),
	convgen.DiscoverFieldsOnly(false, true),
	convgen.DiscoverUnexported(false, true),
	convgen.RenameToUpper(true, true),
)

var SetterOnly = convgen.Struct[X, Y](convgen.Module(
	convgen.DiscoverGetters("Get", ""),
	convgen.DiscoverSetters("Set", ""),
),
	convgen.DiscoverFieldsOnly(true, false),
	convgen.DiscoverUnexported(true, false),
	convgen.RenameToUpper(true, true),
)

var FieldsOnly = convgen.Struct[X, Y](convgen.Module(
	convgen.DiscoverGetters("Get", ""),
	convgen.DiscoverSetters("Set", ""),
),
	convgen.DiscoverFieldsOnly(true, true),
	convgen.DiscoverUnexported(true, true),
)

func main() {
	// Output: {hello 42} (double reversal)
	fmt.Println(GetterSetter(X{name: "hello", Age: 42}))

	// Output: {olleh 42} (single reversal)
	fmt.Println(GetterOnly(X{name: "hello", Age: 42}))

	// Output: {olleh 42} (single reversal)
	fmt.Println(SetterOnly(X{name: "hello", Age: 42}))

	// Output: {hello 42} (no reversal)
	fmt.Println(FieldsOnly(X{name: "hello", Age: 42}))
}
