package controllers

import (
	"context"
	"fmt"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/revel/revel"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"k8s-devops-console/app"
	"k8s-devops-console/app/models"
	"net/http"
	"strings"
)

type ApiSettings struct {
	ApiBase

	vaultClient *keyvault.BaseClient
}

type SettingsOverall struct {
	Settings models.AppConfigSettings `json:"Configuration"`
	User map[string]string `json:"User"`
	Team map[string]map[string]string `json:"Team"`
}

func (c ApiSettings) accessCheck() (result revel.Result) {
	return c.ApiBase.accessCheck()
}

func (c ApiSettings) Get() revel.Result {
	var ret SettingsOverall

	ret.Settings = app.AppConfig.Settings
	ret.User = map[string]string{}
	ret.Team = map[string]map[string]string{}

	for _, setting := range app.AppConfig.Settings.User {
		ret.User[setting.Name] = c.getKeyvaultSecret(c.userSecretName(setting.Name))
	}

	for _, team := range c.getUser().Teams {
		ret.Team[team.Name] = map[string]string{}
		for _, setting := range app.AppConfig.Settings.Team {
			ret.Team[team.Name][setting.Name] = c.getKeyvaultSecret(c.teamSecretName(team.Name, setting.Name))
		}
	}

	return c.RenderJSON(ret)
}

func (c ApiSettings) UpdateUser() revel.Result {
	var err error
	result := struct {
		Message string
	} {
		Message: "",
	}

	config := map[string]string{}
	c.Params.Bind(&config, "config")

	// validation
	validationMessages := []string{}
	for _, setting := range app.AppConfig.Settings.User {
		if val, ok := config[setting.Name]; ok {
			if !setting.Validation.Validate(val) {
				validationMessages = append(validationMessages, fmt.Sprintf("Validation of \"%s\" failed (%v)", setting.Label, setting.Validation.HumanizeString()))
			}
		}
	}

	if len(validationMessages) >= 1 {
		result.Message = strings.Join(validationMessages, "\n")
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(result)
	}

	// set values
	for _, setting := range app.AppConfig.Settings.User {
		if val, ok := config[setting.Name]; ok {
			err = c.setKeyvaultSecret(
				c.userSecretName(setting.Name),
				val,
			)

			if err != nil {
				result.Message = fmt.Sprintf("Failed setting keyvault")
				c.Response.Status = http.StatusForbidden
				return c.RenderJSON(result)
			}
		}
	}


	return c.RenderJSON(result)
}

func (c ApiSettings) UpdateTeam(team string) revel.Result {
	var err error
	result := struct {
		Message string
	} {
		Message: "",
	}

	// membership check
	if !c.checkTeamMembership(team) {
		result.Message = fmt.Sprintf("Access to team \"%s\" denied", team)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	config := map[string]string{}
	c.Params.Bind(&config, "config")

	// validation
	validationMessages := []string{}
	for _, setting := range app.AppConfig.Settings.Team {
		if val, ok := config[setting.Name]; ok {
			if !setting.Validation.Validate(val) {
				validationMessages = append(validationMessages, fmt.Sprintf("Validation of \"%s\" failed (%v)", setting.Label, setting.Validation.HumanizeString()))
			}
		}
	}

	if len(validationMessages) >= 1 {
		result.Message = strings.Join(validationMessages, "\n")
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(result)
	}

	// set values
	for _, setting := range app.AppConfig.Settings.Team {
		if val, ok := config[setting.Name]; ok {
			err = c.setKeyvaultSecret(
				c.teamSecretName(team, setting.Name),
				val,
			)

			if err != nil {
				result.Message = fmt.Sprintf("Failed setting keyvault")
				c.Response.Status = http.StatusForbidden
				return c.RenderJSON(result)
			}
		}

	}

	return c.RenderJSON(result)
}

func (c ApiSettings) userSecretName(name string) string {
	return fmt.Sprintf("user---%s---%s", c.getUser().Username, name)
}

func (c ApiSettings) teamSecretName(team, name string) string {
	return fmt.Sprintf("team---%s---%s", team, name)
}

func (c ApiSettings) getKeyvaultClient(vaultUrl string) (*keyvault.BaseClient){
	var err error
	var keyvaultAuth autorest.Authorizer

	if c.vaultClient == nil {

		keyvaultAuth, err = auth.NewAuthorizerFromEnvironmentWithResource("https://vault.azure.net")
		if err != nil {
			panic(err)
		}

		client := keyvault.New()
		client.Authorizer = keyvaultAuth

		c.vaultClient = &client
	}

	return c.vaultClient
}

func (c ApiSettings) setKeyvaultSecret(secretName, secretValue string) (error) {
	ctx := context.Background()
	vaultUrl := app.GetConfigString("azure.vault.url", "")

	enabled := secretValue != ""

	secretName = strings.Replace(secretName, "_", "-", -1)
	secretParamSet := keyvault.SecretSetParameters{}
	secretParamSet.Value = &secretValue

	secretAttributs := keyvault.SecretAttributes{}
	secretAttributs.Enabled = &enabled
	secretParamSet.SecretAttributes = &secretAttributs

	client := c.getKeyvaultClient("")
	_, err := client.SetSecret(ctx, vaultUrl, secretName, secretParamSet)

	return err
}

func (c ApiSettings) getKeyvaultSecret(secretName string) (secretValue string) {
	var err error
	var secretBundle keyvault.SecretBundle
	ctx := context.Background()
	vaultUrl := app.GetConfigString("azure.vault.url", "")

	client := c.getKeyvaultClient("")
	secretBundle, err = client.GetSecret(ctx, vaultUrl, secretName, "")

	if err == nil {
		secretValue = *secretBundle.Value
	}

	return
}
