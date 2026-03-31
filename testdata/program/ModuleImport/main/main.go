//go:build convgen

package main

import (
	"fmt"
	"strconv"

	"github.com/sublee/convgen"
)

type User struct {
	Id   int
	Name string
}

type UserDTO struct {
	ID   string
	NAME string
}

func IntToString(i int) string {
	return strconv.Itoa(i)
}

var mod1 = convgen.Module(
	convgen.ImportFunc(IntToString),
)

var mod2 = convgen.Module(
	convgen.ImportModule(mod1),
	convgen.RenameToLower(true, true),
)

var UserToDTO = convgen.Struct[User, UserDTO](mod2)

func main() {
	dto := UserToDTO(User{
		Id:   42,
		Name: "Alice",
	})
	fmt.Println(dto.ID, dto.NAME)
}
