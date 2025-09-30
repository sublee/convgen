//go:build convgen && !negvnoc

package main

import "github.com/sublee/convgen"

var And = convgen.Struct[struct{}, *struct{}](nil)
