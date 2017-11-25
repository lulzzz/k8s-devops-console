package controllers

import (
	"github.com/revel/revel"
	"k8s-management/app/routes"
)

type Index struct {
	Base
}

func (c Index) Home() revel.Result {
	if c.getUser() == nil {
		return c.Render()
	} else {
		return c.Redirect(routes.App.Namespace())
	}
}

func (c Index) Login(username, password string) revel.Result {
	if username == "admin" && password == "admin" {
		c.Session["user"] = username
		return c.Redirect(routes.App.Namespace())
	} else {
		c.Flash.Error("Username or password wrong, retry again")
		return c.Redirect(routes.Index.Home())
	}
}

func (c Index) Logout() revel.Result {
	for k := range c.Session {
		delete(c.Session, k)
	}
	return c.Redirect(routes.Index.Home())
}
