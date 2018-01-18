package controllers

import (
	"fmt"
	"regexp"
	"strings"
	"net/http"
	"github.com/revel/revel"
	"k8s-devops-console/app"
	"k8s-devops-console/app/models"
	"k8s.io/api/core/v1"
)

type Base struct {
	*revel.Controller
}

type userSessionStruct struct {
	User string `json:"u"`
	Id string  `json:"id"`
	Groups []string  `json:"g"`
}

func (c Base) accessCheck() (result revel.Result) {
	if c.getUser() == nil {
		c.Response.Status = http.StatusForbidden
		result = c.RenderTemplate("Home/Index.html")
	}
	return
}

func (c Base) setUser(user models.User) {
	// call session
	c.ViewArgs["user"] = user

	// cookie app version
	c.Session["version"] = app.AppVersion

	// cookie session
	c.Session["user"], _ = user.ToJson()
}

func (c Base) getUser() (user *models.User) {
	// call session
	if c.ViewArgs["user"] != nil {
		user = c.ViewArgs["user"].(*models.User)
		return
	}

	if cookieAppVersion, ok := c.Session["version"]; ok {
		if cookieAppVersion != app.AppVersion {
			// force logout, wrong version
			for k := range c.Session {
				delete(c.Session, k)
			}
			return
		}
	}

	// cookie session
	if jsonVal, ok := c.Session["user"]; ok {
		newUser, err := models.UserCreateFromJson(jsonVal, app.AppConfig)
		if err == nil {
			user = newUser
			c.ViewArgs["user"] = newUser
		}
	}

	return
}

func (c Base) checkTeamMembership(teamName string) (status bool) {
	status = false

	for _, team := range c.getUser().Teams {
		if teamName == team.Name {
			status = true
			break
		}
	}

	return
}

func (c Base) checkKubernetesNamespaceAccess(namespace v1.Namespace) (bool) {
	user := c.getUser();

	username := strings.ToLower(user.Username)
	username = strings.Replace(username, "_", "", -1)

	// USER namespace
	regexpUser := regexp.MustCompile(fmt.Sprintf(app.NamespaceFilterUser, regexp.QuoteMeta(username)));
	if regexpUser.MatchString(namespace.Name) {
		return true
	}

	labelUserKey := app.GetConfigString("k8s.label.user", "user");
	labelTeamKey := app.GetConfigString("k8s.label.team", "team");

	if val, ok := namespace.Labels[labelUserKey]; ok {
		if val == user.Username {
			return true
		}
	}

	// ENV namespace (team labels)
	for _, team := range user.Teams {
		if val, ok := namespace.Labels[labelTeamKey]; ok {
			if val == team.Name {
				return true
			}
		}
	}

	// TEAM namespace
	teamsQuoted := []string{}
	for _, team := range user.Teams {
		teamsQuoted = append(teamsQuoted, regexp.QuoteMeta(team.Name))
	}

	regexpTeamStr := fmt.Sprintf(app.NamespaceFilterTeam, "(" + strings.Join(teamsQuoted, "|") + ")")
	regexpTeam := regexp.MustCompile(regexpTeamStr)
	if regexpTeam.MatchString(namespace.Name) {
		return true
	}


	return false
}

func (c Base) renderJSONError(err string) (revel.Result) {
	c.Response.Status = http.StatusInternalServerError
	result := struct {
		Message string
	}{
		Message: fmt.Sprintf("Error: %v", err),
	}
	return c.RenderJSON(result)
}

func (c Base) auditLog(msg string, ctx ...interface{}) {
	msg = fmt.Sprintf("User(%s): %s", c.getUser().Username, msg)
	app.AuditLog.Info(msg)
}
