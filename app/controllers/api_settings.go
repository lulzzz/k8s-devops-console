package controllers

import (
	"context"
	"fmt"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/revel/revel"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"k8s-devops-console/app"
	"net/http"
	"strings"
)

type ApiSettings struct {
	ApiBase

	vaultClient *keyvault.BaseClient
}

type SettingsOverall struct {
	Personal SettingsPersonal `json:"personal"`
	Team map[string]SettingsTeam `json:"team"`
}

type SettingsPersonal struct {
	SshPubKey string
}

type SettingsTeam struct {
	AlertingSlackApi string
	AlertingPagerdutyApi string
}

func (c ApiSettings) accessCheck() (result revel.Result) {
	return c.ApiBase.accessCheck()
}

func (c ApiSettings) Get() revel.Result {
	var ret SettingsOverall
	ret.Team = map[string]SettingsTeam{}

	ret.Personal.SshPubKey = c.getKeyvaultSecret(c.personalSecretName("SshPubKey"))

	for _, team := range c.getUser().Teams {
		ret.Team[team.Name] = SettingsTeam{
			AlertingSlackApi: c.getKeyvaultSecret(c.teamSecretName(team.Name, "AlertingSlackApi")),
			AlertingPagerdutyApi: c.getKeyvaultSecret(c.teamSecretName(team.Name, "AlertingPagerdutyApi")),
		}
	}

	return c.RenderJSON(ret)
}

func (c ApiSettings) UpdatePersonal() revel.Result {
	var err error
	result := struct {
		Message string
	} {
		Message: "",
	}

	config := map[string]string{}
	c.Params.Bind(&config, "config")

	// ssh pub key
	if val, ok := config["SshPubKey"]; ok {
		err = c.setKeyvaultSecret(
			c.personalSecretName("SshPubKey"),
			val,
		)

		if err != nil {
			result.Message = fmt.Sprintf("Failed setting keyvault")
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
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

	// alerting slackapi
	if val, ok := config["AlertingSlackApi"]; ok {
		err = c.setKeyvaultSecret(
			c.teamSecretName(team, "AlertingSlackApi"),
			val,
		)

		if err != nil {
			result.Message = fmt.Sprintf("Failed setting keyvault")
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}
	}

	// alerting pagerdutyapi
	if val, ok := config["AlertingPagerdutyApi"]; ok {
		err = c.setKeyvaultSecret(
			c.teamSecretName(team, "AlertingPagerdutyApi"),
			val,
		)

		if err != nil {
			result.Message = fmt.Sprintf("Failed setting keyvault")
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}
	}


	return c.RenderJSON(result)
}

func (c ApiSettings) personalSecretName(name string) string {
	return fmt.Sprintf("personal---%s---%s", c.getUser().Username, name)
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
