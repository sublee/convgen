package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("# Case 1: Successful conversion")
	fmt.Println()

	u1 := User{
		ID:         unique(1234567890),
		Name:       "John Doe",
		URLs:       []string{"https://example.com"},
		Role:       UserRoleMember,
		CreateTime: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC),
	}
	fmt.Println("Input:")
	fmt.Printf("\t%+v\n", u1)

	fmt.Println("Convgen:")
	cgOut, _ := ConvgenVersion(u1)
	fmt.Printf("\t%s\n", cgOut)

	fmt.Println("Goverter:")
	gvOut, _ := GoverterVersion(u1)
	fmt.Printf("\t%s\n", gvOut)

	fmt.Println()

	fmt.Println("# Case 2: Comprehensive error message")
	fmt.Println()

	u2 := User{Name: "Alice"}
	fmt.Println("Input:")
	fmt.Printf("\t%+v\n", u2)

	fmt.Println("Convgen:")
	_, cgErr := ConvgenVersion(u2)
	fmt.Printf("\t%s\n", cgErr.Error())

	fmt.Println("Goverter:")
	_, gvErr := GoverterVersion(u2)
	fmt.Printf("\t%s\n", gvErr.Error())
}
