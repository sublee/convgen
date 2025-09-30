//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type (
	X  struct{ Child XX }
	Y  struct{ TheChild YY }
	XX struct{ THEME string }
	YY struct{ Theme string }
)

var (
	mod   = convgen.Module(convgen.RenameToUpper(true, true))
	refer = convgen.Struct[X, Y](
		// Affect to all fields including nested fields.
		mod,

		// Not affect to nested fields. If affected, it would rename THEME to ME and
		// cause match error.
		convgen.RenameReset(true, true),
		convgen.RenameTrimPrefix("", "The"),
	)
)

var inline = convgen.Struct[X, Y](
	// Affect to all fields including nested fields.
	convgen.Module(convgen.RenameToUpper(true, true)),

	// Not affect to nested fields. If affected, it would rename THEME to ME and
	// cause match error.
	convgen.RenameReset(true, true),
	convgen.RenameTrimPrefix("", "The"),
)

func main() {
	x := X{Child: XX{THEME: "ok"}}

	// Output: ok ok
	fmt.Println(refer(x).TheChild.Theme)
	fmt.Println(inline(x).TheChild.Theme)
}
