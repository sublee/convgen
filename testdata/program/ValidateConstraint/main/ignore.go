//go:build ignore

package main

import "github.com/sublee/convgen"

var Ignore = convgen.Struct[struct{}, *struct{}](nil)
