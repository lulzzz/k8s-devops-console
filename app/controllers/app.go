package controllers

import (
	"github.com/revel/revel"
	"k8s-devops-console/app"
)

type App struct {
	Base
}

func (c App) accessCheck() (result revel.Result) {
	return c.Base.accessCheck()
}

func (c App) User() revel.Result {
	return c.Render()
}

func (c App) Cluster() revel.Result {
	return c.Render()
}

func (c App) Namespace() revel.Result {
	c.ViewArgs["namespaceEnvironments"] = app.NamespaceEnvironments
	return c.Render()
}

func (c App) About() revel.Result {
	return c.Render()
}
