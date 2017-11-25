package controllers

import (
	"fmt"
	"github.com/revel/revel"
	"k8s.io/api/core/v1"
	"k8s-management/app/services"
)

type Ajax struct {
	Base
}

func (c Ajax) accessCheck() (result revel.Result) {
	return c.Base.accessCheck();
}

func (c Ajax) Cluster() revel.Result {
	service := services.Kubernetes{}
	nodes, err := service.Nodes()

	if err == nil {
		c.ViewArgs["nodes"] = nodes.Items
	} else {
		c.Log.Error(fmt.Sprintf("K8S error: %v", err))
		c.Flash.Error(fmt.Sprintf("Communcation error: %v", err))
	}

	return c.Render()
}

func (c Ajax) Namespace() revel.Result {
	service := services.Kubernetes{}
	nsList, err := service.NamespaceList();

	namespaceList := map[string]v1.Namespace{}
	for nsName, nsObject := range nsList {
		if c.checkKubernetesNamespaceAccess(nsObject) {
			namespaceList[nsName] = nsObject
		}
	}

	if err == nil {
		c.ViewArgs["namespaces"] = namespaceList
	} else {
		c.Log.Error(fmt.Sprintf("K8S error: %v", err))
		c.Flash.Error(fmt.Sprintf("Communcation error: %v", err))
	}

	return c.Render()
}
