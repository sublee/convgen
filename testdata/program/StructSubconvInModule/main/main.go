//go:build convgen

package main

import (
	"fmt"
	"runtime"

	"github.com/sublee/convgen"
)

type (
	XX struct{ I int }
	YY struct{ I int }
)

type (
	X1 struct{ Sub XX }
	Y1 struct{ Sub YY }
	X2 struct{ Sub XX }
	Y2 struct{ Sub YY }
)

var recordedLines = make(map[int]struct{})

func recordCallerLine(i int) int {
	_, _, line, _ := runtime.Caller(1)
	recordedLines[line] = struct{}{}
	return i
}

var mod = convgen.Module(
	convgen.ImportFunc(recordCallerLine),
)

// Both should share the XX -> YY subconv. This can be verified by the length of
// recordedLines. If they share correctly, the length will be 1 rather than 2.
var (
	XtoY1 = convgen.Struct[X1, Y1](mod)
	XtoY2 = convgen.Struct[X2, Y2](mod)
)

func main() {
	// Output: 1 2
	fmt.Println(XtoY1(X1{Sub: XX{I: 1}}).Sub.I)
	fmt.Println(XtoY2(X2{Sub: XX{I: 2}}).Sub.I)

	// Output: true
	fmt.Println(len(recordedLines) == 1)
}
