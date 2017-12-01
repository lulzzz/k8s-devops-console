package controllers

import (
	"fmt"
	"math/rand"
	"encoding/base64"
	"github.com/revel/revel"
	"k8s-devops-console/app/routes"
	"k8s-devops-console/app/models"
	"k8s-devops-console/app/services"
)

type Home struct {
	Base
}

func (c Home) Index() revel.Result {
	if c.getUser() == nil {
		return c.Render()
	} else {
		return c.Redirect(routes.App.Namespace())
	}
}

func (c Home) OAuthStart(username, password string) revel.Result {
	b := make([]byte, 16)
	rand.Read(b)

	state := base64.URLEncoding.EncodeToString(b)

	c.Session["oauth"] = state

	oauth := services.OAuth{}
	url := oauth.AuthCodeURL(state)
	return c.Redirect(url)
}

func (c Home) OAuthAuthorize() revel.Result {
	var user models.User
	oauth := services.OAuth{}

	code := c.Params.Query.Get("code")
	if code == "" {
		c.Flash.Error("OAuth pre check failed: code empty")
		return c.Redirect(routes.Home.Index())
	}

	state := c.Params.Query.Get("state")
	if state == "" {
		c.Flash.Error("OAuth pre check failed: state empty")
		return c.Redirect(routes.Home.Index())
	}

	if state != c.Session["oauth"] {
		c.Flash.Error("OAuth pre check failed: state mismatch")
		return c.Redirect(routes.Home.Index())
	}

	tkn, err := oauth.Exchange(code)
	if err != nil {
		c.Log.Error(fmt.Sprintf("OAUTH Exchange error: %v",err))
		c.Flash.Error("OAuth failed: failed getting token from provider")
		return c.Redirect(routes.Home.Index())
	}

	if !tkn.Valid() {
		c.Flash.Error("OAuth failed: invalid token")
		return c.Redirect(routes.Home.Index())
	}


	user, err = oauth.FetchUserInfo(tkn)
	if err != nil {
		c.Log.Error(fmt.Sprintf("OAUTH fetch user error: %v",err))
		c.Flash.Error("OAuth failed: failed to get user information")
		return c.Redirect(routes.Home.Index())
	}

	c.setUser(user)

	return c.Redirect(routes.App.Namespace())
}

func (c Home) Logout() revel.Result {
	for k := range c.Session {
		delete(c.Session, k)
	}
	return c.Redirect(routes.Home.Index())
}
