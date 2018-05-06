package models

import (
	"fmt"
)

type TeamPermissionsServiceAccount struct {
	Name string  `yaml:"name"`
	Namespace string  `yaml:"namespace"`
}

type TeamPermissions struct {
	Name string  `yaml:"name"`
	Groups []string  `yaml:"groups"`
	Users []string  `yaml:"users"`
	ServiceAccounts []TeamPermissionsServiceAccount  `yaml:"serviceaccounts"`
	ClusterRole string  `yaml:"clusterrole"`
}

type Team struct {
	Name string `json:"-"`
	Permissions []TeamPermissions `json:"-"`
}

func (t *Team) String() string {
	return fmt.Sprintf("Team(%s)", t.Name)
}
