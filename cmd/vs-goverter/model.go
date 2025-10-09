package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type User struct {
	ID         unique
	Name       string
	URLs       []string
	Role       UserRole
	CreateTime time.Time
}

type unique int64

func (u unique) String() string { return strconv.FormatInt(int64(u), 16) }

func firstname(name string) (string, error) {
	parts := strings.Split(name, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("need two parts to parse firstname: %q", name)
	}
	return parts[0], nil
}

func lastname(name string) (string, error) {
	parts := strings.Split(name, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("need two parts to parse firstname: %q", name)
	}
	return parts[1], nil
}

type UserRole int

const (
	UserRoleUnknown UserRole = iota
	UserRoleAdmin
	UserRoleMember
	UserRoleGuest
)
