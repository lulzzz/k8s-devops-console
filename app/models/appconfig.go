package models

import (
	yaml "gopkg.in/yaml.v2"
)

type AppConfigDefault struct {
	Teams []string `yaml:"teams"`
}

type AppConfigUser struct {
	Teams []string `yaml:"teams"`
}

type AppConfigGroup struct {
	Teams []string `yaml:"teams"`
}

type AppConfigTeam struct {
	RoleBinding []TeamPermissions `yaml:"rolebinding"`
}

type AppConfig struct {
	Default AppConfigDefault `yaml:"default"`
	User map[string]AppConfigUser `yaml:"user"`
	Group map[string]AppConfigGroup `yaml:"group"`
	Team map[string]AppConfigTeam `yaml:"team"`
}

func AppConfigCreateFromYaml(yamlString string) (c *AppConfig, err error) {
	err = yaml.Unmarshal([]byte(yamlString), &c)
	return
}
