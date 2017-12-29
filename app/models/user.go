package models

import (
	"fmt"
)

type User struct {
	Id       string
	Username string
	Email    string
	Teams    []Team
	IsAdmin  bool
}

func (u *User) String() string {
	return fmt.Sprintf("User(%s)", u.Username)
}
