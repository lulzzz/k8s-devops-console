package controllers

import (
	"github.com/revel/revel"
)

type ApiUser struct {
	ApiBase
}

func (c ApiUser) accessCheck() (result revel.Result) {
	return c.ApiBase.accessCheck()
}

func (c ApiUser) Kubeconfig() revel.Result {
	c.Response.ContentType = "text/yaml"
	c.Response.Out.Header().Set("Content-Disposition", "attachment; filename=\"kubeconfig.yaml\"")
	return c.Render()
}
