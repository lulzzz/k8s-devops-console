package models

import (
	"fmt"
)

type Team struct {
	Name               string
}

func (t *Team) String() string {
	return fmt.Sprintf("Team(%s)", t.Name)
}