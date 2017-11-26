package controllers

import (
	"github.com/revel/revel"
	"net/http"
)

type ApiLogin struct {
	Base
}

func (c ApiLogin) accessCheck() (result revel.Result) {
	return nil
}

func (c ApiLogin) Login(username, password string) revel.Result {
	if username == "admin" && password == "admin" {
		c.Session["user"] = username
		return c.RenderJSON(true)
	}

	c.Response.Status = http.StatusForbidden
	return c.RenderJSON(false)
}