package models

import (
	yaml "gopkg.in/yaml.v2"
)

type AppConfigUser struct {
	Teams []string `yaml:"teams"`
}

type AppConfigGroup struct {
	Teams []string `yaml:"teams"`
}

type AppConfigTeamRoleBinding struct {
	Name string `yaml:"name"`
	Groups []string `yaml:"groups"`
	ClusterRole string `yaml:"clusterrole"`
}

type AppConfigTeam struct {
	RoleBinding []AppConfigTeamRoleBinding `yaml:"rolebinding"`
}

type AppConfig struct {
	User map[string]AppConfigUser `yaml:"user"`
	Group map[string]AppConfigGroup `yaml:"group"`
	Team map[string]AppConfigTeam `yaml:"team"`
}

func AppConfigCreateFromYaml(yamlString string) (c *AppConfig, err error) {
	err = yaml.Unmarshal([]byte(yamlString), &c)
	return
}
