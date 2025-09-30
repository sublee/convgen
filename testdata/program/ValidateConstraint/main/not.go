//go:build !convgen

package main

import "github.com/sublee/convgen"

var Not = convgen.Struct[struct{}, *struct{}](nil)
