package services

import (
	"fmt"
	"context"
	"github.com/revel/revel"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	githubapi "github.com/google/go-github/github"
	"k8s-devops-console/app/models"
)

var (
	OAuthProvider string
)

type OAuth struct {
	config *oauth2.Config
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
	client := o.GetConfig().Client(context.Background(), token)

	switch o.provider {
	case "github":
		client := githubapi.NewClient(client)
		githubUser, _, err := client.Users.Get(context.Background(), "")
		if err != nil {
			error = err
			return
		}
		user.Username = githubUser.GetLogin()
		user.Email = githubUser.GetEmail()
	default:
		panic(fmt.Sprintf("oauth.provider \"%s\" is not valid", OAuthProvider))
	}

	return
}

func (o *OAuth) buildConfig() (config *oauth2.Config) {
	var clientId, clientSecret string
	var optExists bool
	var endpoint oauth2.Endpoint

	o.provider, optExists = revel.Config.String("oauth.provider")
	if !optExists {
		panic("No oauth.provider configured")
	}

	switch o.provider {
	case "github":
		endpoint = github.Endpoint
	default:
		panic(fmt.Sprintf("oauth.provider \"%s\" is not valid", OAuthProvider))
	}

	if val, exists := revel.Config.String("oauth.endpoint.auth"); exists && val != "" {
		endpoint.AuthURL = val
	}

	if val, exists := revel.Config.String("oauth.endpoint.token"); exists && val != "" {
		endpoint.TokenURL = val
	}

	clientId, optExists = revel.Config.String("oauth.client.id")
	if !optExists {
		panic("No oauth.client.id configured")
	}

	clientSecret, optExists = revel.Config.String("oauth.client.secret")
	if !optExists {
		panic("No oauth.client.secret configured")
	}

	config = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint: endpoint,
		Scopes: []string{},
	}

	return
}
