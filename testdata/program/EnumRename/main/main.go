//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type Status int

const (
	Pending Status = iota
	InProgress
	Done
)

type STATUS string

const (
	STATUS_UNSPECIFIED STATUS = ""
	PENDING            STATUS = "PENDING"
	IN_PROGRESS        STATUS = "IN_PROGRESS"
	DONE               STATUS = "DONE"
)

var ToUpperSnakeCase = convgen.Enum[Status, STATUS](nil, STATUS_UNSPECIFIED,
	convgen.RenameToUpper(true, false),
	convgen.RenameReplace("", "", "_", ""),
)

func main() {
	// Output: PENDING IN_PROGRESS DONE
	fmt.Println(ToUpperSnakeCase(Pending), ToUpperSnakeCase(InProgress), ToUpperSnakeCase(Done))
}
