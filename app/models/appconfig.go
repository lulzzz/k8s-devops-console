package models

import (
	yaml "gopkg.in/yaml.v2"
)

type AppConfig struct {
	Settings AppConfigSettings `yaml:"settings"`
	Permissions AppConfigPermissions `yaml:"permissions"`
}

type AppConfigPermissions struct {
	Default AppConfigDefault `yaml:"default"`
	User map[string]AppConfigUser `yaml:"user"`
	Group map[string]AppConfigGroup `yaml:"group"`
	Team map[string]AppConfigTeam `yaml:"team"`
}

type AppConfigSettings struct {
	User []AppConfigSettingItem
	Team []AppConfigSettingItem
}

type AppConfigSettingItem struct {
	Name string
	Label string
	Type string
	Placeholder string
	Validation struct {
		Regexp string
	}
}

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
	K8sRoleBinding []TeamK8sPermissions `yaml:"rolebinding"`
	AzureRoleAssignments []TeamAzureRoleAssignments `yaml:"azureroleassignment"`
}

func AppConfigCreateFromYaml(yamlString string) (c *AppConfig, err error) {
	err = yaml.Unmarshal([]byte(yamlString), &c)
	return
}
