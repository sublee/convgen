//go:build convgen

package main

import "example.com/SameImportName/foo/pkg"

func init() {
	pkg.Foo()
}
