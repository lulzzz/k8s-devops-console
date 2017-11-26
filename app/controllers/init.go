
package controllers

import "github.com/revel/revel"

func init() {
	revel.InterceptMethod(App.accessCheck, revel.BEFORE)
	revel.InterceptMethod(ApiCluster.accessCheck, revel.BEFORE)
	revel.InterceptMethod(ApiNamespace.accessCheck, revel.BEFORE)
	revel.InterceptMethod(ApiUser.accessCheck, revel.BEFORE)
}