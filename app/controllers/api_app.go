package controllers

import (
	"github.com/revel/revel"
	"k8s-devops-console/app"
)

type ResultConfig struct {
	User ResultUser
	Teams []ResultTeam
	NamespaceEnvironments []string
	Quota map[string]int
}

type ResultUser struct {
	Name string
	Username string
}

type ResultTeam struct {
	Id string
	Name string
}

type ApiApp struct {
	ApiBase
}

func (c ApiApp) accessCheck() (result revel.Result) {
	return c.ApiBase.accessCheck()
}

func (c ApiApp) Config() revel.Result {
	ret := ResultConfig{}
	ret.User.Username = c.getUser().Username

	for _, team := range c.getUser().Teams {
		row := ResultTeam{
			Id: team.Name,
			Name: team.Name,
		}
		ret.Teams = append(ret.Teams, row)
	}

	ret.NamespaceEnvironments = app.NamespaceEnvironments


	ret.Quota = map[string]int{
		"team": app.GetConfigInt("k8s.namespace.team.quota", 0),
		"user": app.GetConfigInt("k8s.namespace.user.quota", 0),
	}

	return c.RenderJSON(ret)
}
