package controllers

import (
	"github.com/revel/revel"
	"k8s-management/app"
)

type ResultConfig struct {
	User ResultUser
	Teams []ResultTeam
	NamespaceEnvironments []string
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
	Base
}

func (c ApiApp) accessCheck() (result revel.Result) {
	return c.Base.accessCheck()
}

func (c ApiApp) Config() revel.Result {
	ret := ResultConfig{}
	ret.User.Name = c.getUser().Name
	ret.User.Username = c.getUser().Username

	for _, team := range c.getUser().Teams {
		row := ResultTeam{
			Id: team.Name,
			Name: team.Name,
		}
		ret.Teams = append(ret.Teams, row)
	}

	ret.NamespaceEnvironments = app.NamespaceEnvironments
	return c.RenderJSON(ret)
}