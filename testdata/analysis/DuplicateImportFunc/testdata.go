//go:build convgen

package testdata

import (
	"strconv"

	"github.com/sublee/convgen"
)

func int2string(int) string { return "" }

func string2int(string) (int, error) { return 0, nil }

func Identity[T any](x T) T { return x }

type (
	TheInt = int
	MyInt  int
)

var _ = convgen.Module(
	// ImportFunc
	convgen.ImportFunc(strconv.Itoa),
	convgen.ImportFunc(int2string),                     // want `duplicate int to string converter`
	convgen.ImportFunc(func(int) string { return "" }), // want `duplicate int to string converter`

	// Generic function instantiations
	convgen.ImportFunc(Identity[string]), // ok
	convgen.ImportFunc(Identity[string]), // want `duplicate string to string converter`
	convgen.ImportFunc(Identity[int]),    // ok

	// Type aliases and defined types
	convgen.ImportFunc(func(TheInt) string { return "" }), // want `duplicate TheInt to string converter`
	convgen.ImportFunc(func(MyInt) string { return "" }),  // ok, because MyInt is different from int

	// ImportFuncErr
	convgen.ImportFuncErr(strconv.Atoi),
	convgen.ImportFuncErr(string2int),                                  // want `duplicate string to int converter`
	convgen.ImportFuncErr(func(string) (int, error) { return 0, nil }), // want `duplicate string to int converter`
)
