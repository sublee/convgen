//go:build convgen

package main

import (
	"fmt"

	"github.com/sublee/convgen"
)

type Foo struct {
	name    string
	address Address
	status  Status
}

type Bar struct {
	name    string
	address PostfixedAddress
	status  PrefixedStatus
}

type (
	Address          interface{ isAddress() }
	PostfixedAddress interface{ isPostfixedAddress() }
)

type (
	Email struct{ Email string }
	GPS   struct{ Lat, Long float64 }
)

type (
	EmailAddress struct{ Email string }
	GPSAddress   struct{ Lat, Long float64 }
)

func (Email) isAddress()                 {}
func (GPS) isAddress()                   {}
func (EmailAddress) isPostfixedAddress() {}
func (GPSAddress) isPostfixedAddress()   {}

type (
	Status         int
	PrefixedStatus string
)

const (
	Pending Status = iota
	InProgress
	Done
)

const (
	StatusUnspecified PrefixedStatus = ""
	StatusPending     PrefixedStatus = "pending"
	StatusInProgress  PrefixedStatus = "in_progress"
	StatusDone        PrefixedStatus = "done"
)

var mod = convgen.Module(
	convgen.ForStruct(convgen.DiscoverUnexported(true, true)),
	convgen.ForUnion(convgen.RenameTrimCommonSuffix(false, true)),
	convgen.ForEnum(convgen.RenameTrimCommonPrefix(false, true)),
)

var (
	FooToBar         = convgen.Struct[Foo, Bar](mod)
	FooToBar_address = convgen.Union[Address, PostfixedAddress](mod)
	FooToBar_status  = convgen.Enum[Status, PrefixedStatus](mod, StatusUnspecified)
)

func main() {
	bar := FooToBar(Foo{
		name:    "Test",
		address: Email{Email: "test@example.com"},
		status:  InProgress,
	})
	fmt.Println(bar.name)
	fmt.Println(bar.address.(EmailAddress).Email)
	fmt.Println(bar.status)
}
