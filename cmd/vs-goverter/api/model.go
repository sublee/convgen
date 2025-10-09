package api

import (
	"encoding/json"
)

type User struct {
	Id        string   `json:"id"`
	Firstname string   `json:"firstname"`
	Lastname  string   `json:"lastname"`
	Urls      []string `json:"urls"`
	Role      UserRole `json:"role"`
	CreatedAt int64    `json:"createdAt"`
}

func (u User) String() string {
	b, _ := json.Marshal(u)
	return string(b)
}

type UserRole int

const (
	UserRoleUnspecified UserRole = iota
	Admin
	Member
	Guest
)

func (r UserRole) MarshalJSON() ([]byte, error) {
	switch r {
	case Admin:
		return json.Marshal("admin")
	case Member:
		return json.Marshal("member")
	case Guest:
		return json.Marshal("guest")
	}
	return json.Marshal(nil)
}
