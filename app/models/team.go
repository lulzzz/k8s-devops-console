package models

import (
	"fmt"
)

type TeamPermissions struct {
	Name string
	Groups []string
	ClusterRole string
}

type Team struct {
	Name string `json:"-"`
	Permissions []TeamPermissions `json:"-"`
}

func (t *Team) String() string {
	return fmt.Sprintf("Team(%s)", t.Name)
}
