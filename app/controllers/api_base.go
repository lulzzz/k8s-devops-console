package controllers

import (
	"errors"
	"github.com/revel/revel"
	"net/http"
)

type ApiBase struct {
	Base
}

func (c ApiBase) accessCheck() (result revel.Result) {
	if c.getUser() == nil {
		c.Response.Status = http.StatusUnauthorized
		result = c.RenderError(errors.New("not logged in"))
	}
	return
}

