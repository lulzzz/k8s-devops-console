package models

import (
	"fmt"
)

type TeamPermissionsServiceAccount struct {
	Name string
	Namespace string
}

type TeamPermissions struct {
	Name string
	Groups []string
	Users []string
	ServiceAccounts []TeamPermissionsServiceAccount
	ClusterRole string
}

type Team struct {
	Name string `json:"-"`
	Permissions []TeamPermissions `json:"-"`
}

func (t *Team) String() string {
	return fmt.Sprintf("Team(%s)", t.Name)
}
