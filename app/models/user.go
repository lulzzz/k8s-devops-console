package models

import (
	"fmt"
)

type User struct {
	UserId             int
	Name               string
	Username, Password string
	Email              string
	Teams              []Team
	IsAdmin            bool
}

func (u *User) String() string {
	return fmt.Sprintf("User(%s)", u.Username)
}
