package controllers

import (
	"github.com/revel/revel"
)

type ApiUser struct {
	Base
}

func (c ApiUser) accessCheck() (result revel.Result) {
	return c.Base.accessCheck()
}

func (c ApiUser) Kubeconfig() revel.Result {
	c.Response.ContentType = "application/json"
	c.Response.Out.Header().Set("Content-Disposition", "attachment; filename=\"kubeconfig.json\"")
	return c.Render()
}
