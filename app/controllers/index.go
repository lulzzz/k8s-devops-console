package controllers

import (
	"fmt"
	"context"
	"math/rand"
	"encoding/base64"
	"github.com/revel/revel"
	"k8s-devops-console/app"
	"k8s-devops-console/app/routes"
	"k8s-devops-console/app/models"
	"github.com/google/go-github/github"
)

type Index struct {
	Base
}

func (c Index) Home() revel.Result {
	if c.getUser() == nil {
		return c.Render()
	} else {
		return c.Redirect(routes.App.Namespace())
	}
}

func (c Index) OAuthStart(username, password string) revel.Result {
	b := make([]byte, 16)
	rand.Read(b)

	state := base64.URLEncoding.EncodeToString(b)

	c.Session["oauth"] = state

	url := app.OAuthConfig.AuthCodeURL(state)
	return c.Redirect(url)
}

func (c Index) OAuthAuthorize() revel.Result {
	code := c.Params.Query.Get("code")
	if code == "" {
		c.Flash.Error("OAuth pre check failed: code empty")
		return c.Redirect(routes.Index.Home())
	}

	state := c.Params.Query.Get("state")
	if state == "" {
		c.Flash.Error("OAuth pre check failed: state empty")
		return c.Redirect(routes.Index.Home())
	}

	if state != c.Session["oauth"] {
		c.Flash.Error("OAuth pre check failed: state mismatch")
		return c.Redirect(routes.Index.Home())
	}

	tkn, err := app.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.Log.Error(fmt.Sprintf("OAUTH Exchange error: %v",err))
		c.Flash.Error("OAuth failed: failed getting token from provider")
		return c.Redirect(routes.Index.Home())
	}

	if !tkn.Valid() {
		c.Flash.Error("OAuth failed: invalid token")
		return c.Redirect(routes.Index.Home())
	}

	switch app.OAuthProvider {
	case "github":
		client := github.NewClient(app.OAuthConfig.Client(context.Background(), tkn))
		githubUser, _, err := client.Users.Get(context.Background(), "")
		if err != nil {
			c.Flash.Error("OAuth failed: failed getting user informations from github")
			return c.Redirect(routes.Index.Home())
		}

		user := models.User{}
		user.Username = *githubUser.Login
		c.setUser(user)
		break;
	default:
		c.Flash.Error("OAuth provider: not valid")
		return c.Redirect(routes.Index.Home())
	}

	return c.Redirect(routes.App.Namespace())
}

func (c Index) Logout() revel.Result {
	for k := range c.Session {
		delete(c.Session, k)
	}
	return c.Redirect(routes.Index.Home())
}
