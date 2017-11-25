package controllers

import (
	"github.com/revel/revel"
)

type ApiUser struct {
	Base
}


func (c ApiUser) accessCheck() (result revel.Result) {
	return c.Base.accessCheck();
}

func (c ApiUser) Credentials() revel.Result {
	return c.Render()
}
