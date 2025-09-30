//go:build convgen

package main

import (
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

	// "", "", "_", "" is what actually gets passed.
	// Convgen should report how names are renamed to help debugging.
	// e.g., missing: InProgress -> ? (renamed to "_I_N_P_R_O_G_R_E_S_S_")
	convgen.RenameReplace("", "_", "_", ""),
)

func main() {
	panic("convgen will fail")
}
