package models

import (
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"regexp"
	"strings"
)

type AppConfig struct {
	Settings AppConfigSettings `yaml:"settings"`
	Azure AppConfigAzure `yaml:"azure"`
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

type AppInputValidation struct {
	Regexp string
}

type AppConfigSettingItem struct {
	Name string
	Label string
	Type string
	Placeholder string
	Validation AppInputValidation
	Tags map[string]string
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

type AppConfigAzure struct {
	ResourceGroup struct {
		Validation AppInputValidation
		Tags []AppConfigAzureResourceGroupTag
	}
}

type AppConfigAzureResourceGroupTag struct {
	Name string
	Label string
	Type string
	Placeholder string
	Validation AppInputValidation
}

func AppConfigCreateFromYaml(yamlString string) (c *AppConfig, err error) {
	err = yaml.Unmarshal([]byte(yamlString), &c)
	return
}

func (v *AppInputValidation) HumanizeString() (ret string) {
	validationList := []string{}

	if v.Regexp != "" {
		validationList = append(validationList, fmt.Sprintf("regexp:%v", v.Regexp))
	}


	if len(validationList) >= 1 {
		ret = strings.Join(validationList, "; ")
	}

	return
}

func (v *AppInputValidation) Validate(value string) (status bool) {
	status = true

	if v.Regexp != "" {
		validationRegexp := regexp.MustCompile(v.Regexp)

		if !validationRegexp.MatchString(value) {
			status = false
		}
	}

	return
}
