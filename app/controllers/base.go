package controllers

import (
	"fmt"
	"regexp"
	"strings"
	"github.com/revel/revel"
	"k8s-devops-console/app/models"
	"k8s.io/api/core/v1"
	"k8s-devops-console/app"
	"net/http"
	"errors"
)

type Base struct {
	*revel.Controller
}

func (c Base) accessCheck() (result revel.Result) {
	if c.getUser() == nil {
		c.Response.Status = http.StatusForbidden
		result = c.RenderError(errors.New("not logged in"))
	}
	return
}

func (c Base) setUser(user models.User) {
	c.ViewArgs["user"] = user
	c.Session["user"] = user.Username
}

func (c Base) getUser() (user *models.User) {
	if c.ViewArgs["user"] != nil {
		user = c.ViewArgs["user"].(*models.User)
	}
	if username, ok := c.Session["user"]; ok {
		teams := []models.Team{}

		if username == "admin" {
			teams = append(teams, models.Team{Name: "admin"})
			teams = append(teams, models.Team{Name: "user"})
		} else {
			teams = append(teams, models.Team{Name: "user"})
		}
		user = &models.User{Username:username, Teams:teams}
	}
	c.ViewArgs["user"] = user
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

	// USER namespace
	regexpUser := regexp.MustCompile(fmt.Sprintf(app.NamespaceFilterUser, regexp.QuoteMeta(user.Username)));
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
	msg = fmt.Sprintf("[AUDIT] User(%s): %s", c.getUser().Username, msg)
	c.Log.Warn(msg, ctx...)
}
