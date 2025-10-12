//go:build convgen

package main

import (
	"fmt"
	"reflect"

	"github.com/sublee/convgen"
)

type PrivateString struct {
	s string
}

type Tester struct {
	Value   PrivateString
	Pointer *PrivateString

	Array [1]PrivateString
	Slice []PrivateString
	Map   map[string]PrivateString

	ArrayPtr [1]*PrivateString
	SlicePtr []*PrivateString
	MapPtr   map[string]*PrivateString
}

type (
	X Tester
	Y Tester
)

var XtoY = convgen.Struct[X, Y](nil)

func main() { // nolint: gocyclo
	x := X{
		Value:   PrivateString{s: "value"},
		Pointer: &PrivateString{s: "pointer"},

		Array: [1]PrivateString{{s: "array0"}},
		Slice: []PrivateString{{s: "slice0"}, {s: "slice1"}},
		Map:   map[string]PrivateString{"key": {s: "mapvalue"}},

		ArrayPtr: [1]*PrivateString{},
		SlicePtr: []*PrivateString{{s: "sliceptr0"}, nil, {s: "sliceptr2"}},
		MapPtr:   map[string]*PrivateString{"key": {s: "mapptrvalue"}},
	}
	y := XtoY(x)

	if x.Value.s != y.Value.s {
		panic("Value")
	}

	if y.Pointer == nil || x.Pointer.s != y.Pointer.s {
		panic("Pointer")
	}
	if x.Pointer == y.Pointer {
		panic("Pointer not copied")
	}

	for i := range x.Array {
		if x.Array[i].s != y.Array[i].s {
			panic("Array element")
		}
	}

	if len(x.Slice) != len(y.Slice) {
		panic("Slice length")
	}
	for i := range x.Slice {
		if x.Slice[i].s != y.Slice[i].s {
			panic("Slice element")
		}
	}
	if reflect.ValueOf(x.Slice).Pointer() == reflect.ValueOf(y.Slice).Pointer() {
		panic("Slice not copied")
	}

	if len(x.Map) != len(y.Map) {
		panic("Map length")
	}
	for k, v := range x.Map {
		v2, ok := y.Map[k]
		if !ok || v.s != v2.s {
			panic("Map element")
		}
	}
	if reflect.ValueOf(x.Map).Pointer() == reflect.ValueOf(y.Map).Pointer() {
		panic("Map not copied")
	}

	for i := range x.ArrayPtr {
		if (x.ArrayPtr[i] == nil) != (y.ArrayPtr[i] == nil) {
			panic("ArrayPtr element nil")
		}
		if x.ArrayPtr[i] != nil && x.ArrayPtr[i].s != y.ArrayPtr[i].s {
			panic("ArrayPtr element value")
		}
	}

	if len(x.SlicePtr) != len(y.SlicePtr) {
		panic("SlicePtr length")
	}
	for i := range x.SlicePtr {
		if (x.SlicePtr[i] == nil) != (y.SlicePtr[i] == nil) {
			panic("SlicePtr element nil")
		}
		if x.SlicePtr[i] != nil && x.SlicePtr[i].s != y.SlicePtr[i].s {
			panic("SlicePtr element value")
		}
	}
	if reflect.ValueOf(x.SlicePtr).Pointer() == reflect.ValueOf(y.SlicePtr).Pointer() {
		panic("SlicePtr not copied")
	}

	if len(x.MapPtr) != len(y.MapPtr) {
		panic("MapPtr length")
	}
	for k, v := range x.MapPtr {
		v2, ok := y.MapPtr[k]
		if !ok {
			panic("MapPtr key missing")
		}
		if (v == nil) != (v2 == nil) {
			panic("MapPtr element nil")
		}
		if v != nil && v.s != v2.s {
			panic("MapPtr element value")
		}
	}
	if reflect.ValueOf(x.MapPtr).Pointer() == reflect.ValueOf(y.MapPtr).Pointer() {
		panic("MapPtr not copied")
	}

	fmt.Println("OK")
}
