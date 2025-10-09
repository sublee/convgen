package testdata

import "github.com/sublee/convgen" // want `file must have "//go:build convgen" constraint when importing convgen`

func F0() {
	_ = convgen.RenameReset(true, true) // wrong but ok, because parser skips files without the build constraint
}
