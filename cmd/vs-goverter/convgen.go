//go:build convgen

package main

import (
	"time"

	"github.com/sublee/convgen"
	"github.com/sublee/convgen/cmd/vs-goverter/api"
)

var mod = convgen.Module(
	// Global glossary
	convgen.RenameReplace("", "", "Id", "ID"),
	convgen.RenameReplace("", "", "Url", "URL"),

	// Custom conversion by inline functions
	convgen.ImportFunc(func(u unique) string { return u.String() }),
	convgen.ImportFunc(func(t time.Time) int64 { return t.Unix() }),
)

var ConvgenVersion = convgen.StructErr[User, api.User](mod,
	// Split Name into Firstname and Lastname by custom functions
	convgen.MatchFuncErr(User{}.Name, api.User{}.Firstname, firstname),
	convgen.MatchFuncErr(User{}.Name, api.User{}.Lastname, lastname),

	// Match manually
	convgen.Match(User{}.CreateTime, api.User{}.CreatedAt),
)

var convgenVersionUserRole = convgen.Enum[UserRole, api.UserRole](mod, api.UserRoleUnspecified,
	convgen.RenameTrimPrefix("UserRole", ""),
	convgen.MatchSkip(UserRoleUnknown, nil),
)

// avoid unused error
var _ = convgenVersionUserRole
