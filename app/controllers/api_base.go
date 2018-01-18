package controllers

import (
	"errors"
	"net/http"
	"github.com/revel/revel"
)

type ApiBase struct {
	Base
}

func (c ApiBase) accessCheck() (result revel.Result) {
	if c.getUser() == nil {
		c.Response.Status = http.StatusForbidden
		result = c.RenderError(errors.New("not logged in"))
	}
	return
}
