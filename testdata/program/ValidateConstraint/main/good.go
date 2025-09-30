//go:build convgen

package main

import "github.com/sublee/convgen"

var Good = convgen.Struct[struct{}, *struct{}](nil)
