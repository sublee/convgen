package main

import (
	"time"

	"github.com/sublee/convgen/cmd/vs-goverter/api"
)

// goverter:variables
// goverter:extend uniqueToString
// goverter:extend timeToUnix
// goverter:matchIgnoreCase
var (
	// goverter:map Name Firstname | firstname
	// goverter:map Name Lastname  | lastname
	// goverter:map CreateTime CreatedAt
	GoverterVersion func(User) (api.User, error)

	// goverter:enum:unknown UserRoleUnspecified
	// goverter:enum:transform regex UserRole(\w+) $1
	// goverter:enum:map UserRoleUnknown UserRoleUnspecified
	GoverterVersionUserRole func(UserRole) api.UserRole
)

func uniqueToString(u unique) string { return u.String() }

func timeToUnix(t time.Time) int64 { return t.Unix() }

var (
	// avoid unused error
	_ = uniqueToString
	_ = timeToUnix
)
