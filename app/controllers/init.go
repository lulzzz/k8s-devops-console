
package controllers

import "github.com/revel/revel"

func init() {
	revel.InterceptMethod(App.accessCheck, revel.BEFORE)
	revel.InterceptMethod(Ajax.accessCheck, revel.BEFORE)
	revel.InterceptMethod(ApiNamespace.accessCheck, revel.BEFORE)
	revel.InterceptMethod(ApiUser.accessCheck, revel.BEFORE)
}