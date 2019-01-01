package controllers

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
	"encoding/base64"
	"github.com/revel/revel"
	"k8s-devops-console/app/routes"
	"k8s-devops-console/app/models"
	"k8s-devops-console/app/services"
	"regexp"
	"k8s-devops-console/app"
)

type Home struct {
	Base
}

func (c Home) Index() revel.Result {
	c.handleOauthErrors()

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

	app.PrometheusActions.With(prometheus.Labels{"scope": "oauth", "type": "start"}).Inc()

	return c.Redirect(url)
}

func (c Home) OAuthAuthorize() revel.Result {
	var user models.User
	oauth := services.OAuth{}

	if c.handleOauthErrors() {
		return c.RenderTemplate("Home/Index.html")
	}

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

	// check username
	if user.Username == "" {
		c.Log.Error("Got empty username, login failed")
		c.Flash.Error("Got empty username, login failed")
		return c.Redirect(routes.Home.Index())
	}

	if filter := app.GetConfigString("oauth.username.filter.whitelist", ""); filter != "" {
		filterRegexp := regexp.MustCompile(filter);

		if ! filterRegexp.MatchString(user.Username) {
			c.Log.Error(fmt.Sprintf("User %s is not allowed to use this application", user.Username))
			c.Flash.Error(fmt.Sprintf("User %s is not allowed to use this application", user.Username))
			return c.Redirect(routes.Home.Index())
		}
	}

	if filter := app.GetConfigString("oauth.username.filter.blacklist", ""); filter != "" {
		filterRegexp := regexp.MustCompile(filter);

		if filterRegexp.MatchString(user.Username) {
			c.Log.Error(fmt.Sprintf("User %s is not allowed to use this application", user.Username))
			c.Flash.Error(fmt.Sprintf("User %s is not allowed to use this application", user.Username))
			return c.Redirect(routes.Home.Index())
		}
	}

	c.setUser(user)

	app.PrometheusActions.With(prometheus.Labels{"scope": "oauth", "type": "login"}).Inc()

	return c.Redirect(routes.App.Namespace())
}

func (c Home) handleOauthErrors() bool {
	// AzureAD error message
	if error := c.Params.Query.Get("error"); error != "" {
		message := error

		if errorDesc := c.Params.Query.Get("error_description"); errorDesc != "" {
			message = fmt.Sprintf("%s:\n%s", error, errorDesc)
		}

		app.PrometheusActions.With(prometheus.Labels{"scope": "oauth", "type": "failed"}).Inc()

		c.Validation.Error(message)
		return true
	}

	return false
}

func (c Home) Logout() revel.Result {
	for k := range c.Session {
		delete(c.Session, k)
	}

	return c.Redirect(routes.Home.Index())
}
