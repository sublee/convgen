//go:build convgen

package testdata

import "github.com/sublee/convgen"

var (
	_    = convgen.Module()                        // ok
	Conv = convgen.Struct[struct{}, struct{}](nil) // ok
)

var (
	ToLower     = convgen.RenameToLower(true, true)            // want `cannot assign RenameToLower to variable`
	ForStruct   = convgen.ForStruct()                          // want `cannot assign ForStruct to variable`
	FieldGetter = convgen.FieldGetter(func() int { return 0 }) // want `cannot assign FieldGetter to variable`
)
