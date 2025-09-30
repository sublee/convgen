//go:build convgen

package main

import "example.com/SameImportName/bar/pkg"

func init() {
	pkg.Bar()
}
