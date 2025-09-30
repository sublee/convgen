//go:build convgen

package main

import (
	"example.com/ModuleInvalid/myconvgen"
	"github.com/sublee/convgen"
)

// Legal: package-level variable
var mod = convgen.Module()

// Legal: package-level pointer variable
var modPtr = &mod

// Legal: reusing module variable
var modReused = mod

// Legal: blank identifier is ignored
var _ = convgen.Module()

func insideFunc() {
	// Illegal: assign to package-level variable
	mod = *modPtr

	// Illegal: declare with ":="
	mod1 := convgen.Module()
	_ = mod1

	// Illegal: assign with "="
	mod1 = convgen.Module()
	_ = mod1

	var (
		// Illegal: declare with "var ="
		mod2 = convgen.Module()
		_    = mod2
	)

	// Illegal: multiple assignment
	mod3, mod4 := convgen.Module(), convgen.Module()
	_ = mod3
	_ = mod4

	// Illegal: blank identifier is also illegal inside function
	_ = convgen.Module()

	// Undefined but not harmful: call without assignment
	convgen.Module()

	// Illegal: these directives must be inlined
	opt := convgen.RenameToLower(true, true)
	field := convgen.FieldGetter(func() int { return 0 })
	_ = opt
	_ = field
}

// Legal: inline module is allowed
var inlineModule = convgen.Struct[struct{}, *struct{}](convgen.Module())

// Illegal: option as variable
var (
	opt    = convgen.RenameToLower(true, true)
	optVar = convgen.Module(opt)
)

// Illegal: option must be convgen function call
var m2 = convgen.Module(myconvgen.RenameToLower(true, true))

func main() {
	panic("convgen will fail")
}
