package services

import (
	"fmt"
	"strings"
	"context"
	"encoding/json"
	"github.com/revel/revel"
	"k8s-devops-console/app"
	"k8s-devops-console/app/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"github.com/coreos/go-oidc"
	githubapi "github.com/google/go-github/github"
)

var (
	OAuthProvider string
)

type OAuth struct {
	config *oauth2.Config
	oidcProvider *oidc.Provider
	provider string
}

func (o *OAuth) GetConfig() (config *oauth2.Config) {
	if o.config == nil {
		o.config = o.buildConfig()
	}
	config = o.config;
	return
}

func (o *OAuth) GetProvider() (string) {
	return o.provider
}

func (o *OAuth) AuthCodeURL(state string) (string) {
	return o.GetConfig().AuthCodeURL(state)
}

func (o *OAuth) Exchange(code string) (*oauth2.Token, error) {
	return o.GetConfig().Exchange(context.Background(), code)
}

func (o *OAuth) FetchUserInfo(token *oauth2.Token) (user models.User, error error) {
	ctx := context.Background()

	client := o.GetConfig().Client(ctx, token)

	switch strings.ToLower(o.provider) {
	case "github":
		client := githubapi.NewClient(client)
		githubUser, _, err := client.Users.Get(context.Background(), "")
		if err != nil {
			error = err
			return
		}

		user.Id = githubUser.GetLogin()
		user.Username = githubUser.GetLogin()
		user.Email = githubUser.GetEmail()
		user.IsAdmin = githubUser.GetSiteAdmin()
	case "azuread":
		tokenSource := oauth2.StaticTokenSource(token)

		// parse basic user info
		userInfo, err := o.oidcProvider.UserInfo(ctx, tokenSource)
		if err != nil {
			error = err
			return
		}

		// get prefixes from configuration
		userPrefix := app.GetConfigString("oauth.username.prefix", "")
		groupsPrefix := app.GetConfigString("oauth.groups.prefix", "")

		// parse custom userinfo
		aadUserInfo := struct {
			Directory  string `json:"iss"`
			DirectoryId  string `json:"tid"`
			UserId     string `json:"oid"`
			Username   string `json:"upn"`
			PrefUsername string `json:"preferred_username"`
			Groups []string `json:"groups"`
		}{}
		if err := userInfo.Claims(&aadUserInfo); err != nil {
			error = err
			return
		}

		// WORKAROUND: azuread groups (json array as string?!)
		var groupList []string
		for _, val := range aadUserInfo.Groups {
			var tmp []interface{}
			if err := json.Unmarshal([]byte(val), &tmp); err == nil {
				for _, groupName := range tmp {
					groupList = append(groupList, groupName.(string))
				}
			} else {
				groupList = append(groupList, val)
			}
		}

		// add prefix
		for i, val := range groupList {
			groupList[i] = groupsPrefix + val
		}
		aadUserInfo.Groups = groupList

		// extract username from email
		split := strings.SplitN(aadUserInfo.Username, "@", 2)

		// Build user object
		user.Id = userPrefix + aadUserInfo.UserId
		user.Username = split[0]
		user.Email = aadUserInfo.Username
		user.Groups = aadUserInfo.Groups
	default:
		o.error(fmt.Sprintf("oauth.provider \"%s\" is not valid", OAuthProvider))
	}

	if user.Id != "" {
		// Init user
		clusterRole := app.GetConfigString("k8s.user.clusterRole", "")
		if clusterRole != "" {
			service := Kubernetes{}

			app.AuditLog.Info(fmt.Sprintf("User(%s): create k8s ClusterRoleBinding for \"%s\" with Role \"%s\"", user.Username, user.Id, clusterRole))
			if _, err := service.ClusterRoleBindingUser(user.Username, user.Id, clusterRole); err != nil {
				o.error(fmt.Sprintf("Unable to create ClusterRoleBinding: %s", err))
			}
		}
	}

	return
}

func (o *OAuth) buildConfig() (config *oauth2.Config) {
	var clientId, clientSecret string
	var optExists bool
	var endpoint oauth2.Endpoint

	ctx := context.Background()

	scopes := []string{}

	o.provider, optExists = revel.Config.String("oauth.provider")
	if !optExists {
		o.error("No oauth.provider configured")
	}

	switch strings.ToLower(o.provider) {
	case "github":
		endpoint = github.Endpoint
	case "azuread":
		aadTenant := "common"
		if val, exists := revel.Config.String("oauth.azuread.tenant"); exists && val != "" {
			aadTenant = val
		}

		provider, err := oidc.NewProvider(ctx, fmt.Sprintf("https://sts.windows.net/%s/", aadTenant))
		//provider, err := oidc.NewProvider(ctx, fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", aadTenant))
		if err != nil {
			o.error(fmt.Sprintf("oauth.provider AzureAD init failed: %s", err))
		}

		o.oidcProvider = provider
		endpoint = provider.Endpoint()
		scopes = []string{oidc.ScopeOpenID, "profile", "email", "offline_access", "groups"}
	default:
		o.error(fmt.Sprintf("oauth.provider \"%s\" is not valid", OAuthProvider))
	}

	if val, exists := revel.Config.String("oauth.endpoint.auth"); exists && val != "" {
		endpoint.AuthURL = val
	}

	if val, exists := revel.Config.String("oauth.endpoint.token"); exists && val != "" {
		endpoint.TokenURL = val
	}

	clientId, optExists = revel.Config.String("oauth.client.id")
	if !optExists {
		o.error("No oauth.client.id configured")
	}

	clientSecret, optExists = revel.Config.String("oauth.client.secret")
	if !optExists {
		o.error("No oauth.client.secret configured")
	}

	config = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint: endpoint,
		Scopes: scopes,
		RedirectURL: app.GetConfigString("oauth.redirect.url", ""),
	}

	return
}

func (o *OAuth) error(message string) {
	revel.AppLog.Error(message)
	panic(message)
}
