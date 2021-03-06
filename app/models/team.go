package models

import (
	"fmt"
)

type TeamPermissionsServiceAccount struct {
	Name string  `yaml:"name"`
	Namespace string  `yaml:"namespace"`
}

type TeamK8sPermissions struct {
	Name string  `yaml:"name"`
	Groups []string  `yaml:"groups"`
	Users []string  `yaml:"users"`
	ServiceAccounts []TeamPermissionsServiceAccount  `yaml:"serviceaccounts"`
	ClusterRole string  `yaml:"clusterrole"`
}


type TeamAzureRoleAssignments struct {
	Uuid string `yaml:"-"`
	PrincipalId string `yaml:"principalid"`
	Role string `yaml:"role"`
}


type Team struct {
	Name string `json:"-"`
	K8sPermissions []TeamK8sPermissions `json:"-"`
	AzureRoleAssignments []TeamAzureRoleAssignments `json:"-"`
}

func (t *Team) String() string {
	return fmt.Sprintf("Team(%s)", t.Name)
}
