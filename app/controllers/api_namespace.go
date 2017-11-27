package controllers

import (
	"fmt"
	"net/http"
	"github.com/revel/revel"
	"k8s-management/app"
	"k8s-management/app/services"
	"k8s.io/api/core/v1"
	"k8s-management/app/toolbox"
)

type ResultNamespace struct {
	Name string
	OwnerTeam string
	OwnerUser string
	Status string
	Created string
	CreatedAgo string
}

type ApiNamespace struct {
	Base
}

func (c ApiNamespace) accessCheck() (result revel.Result) {
	return c.Base.accessCheck()
}

func (c ApiNamespace) List() revel.Result {
	service := services.Kubernetes{}
	nsList, err := service.NamespaceList()
	if err != nil {
		message := fmt.Sprintf("Error: %v", err)
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJSON(message)
	}

	ret := []ResultNamespace{}

	for _, ns := range nsList {
		row := ResultNamespace{
			Name: ns.Name,
			Status: fmt.Sprintf("%v", ns.Status.Phase),
			Created: ns.CreationTimestamp.UTC().String(),
			CreatedAgo: revel.TimeAgo(ns.CreationTimestamp.UTC()),
		};

		if val, ok := ns.Labels["team"]; ok {
			row.OwnerTeam = val
		}

		if val, ok := ns.Labels["user"]; ok {
			row.OwnerUser = val
		}

		ret = append(ret, row)
	}

	return c.RenderJSON(ret)
}

func (c ApiNamespace) Create(nsEnvironment, nsAreaTeam, nsApp string) revel.Result {
	result := struct {
		Namespace string
		Message string
	} {
		Namespace: "",
		Message: "",
	}

	username := c.getUser().Username

	if ! app.RegexpNamespaceApp.MatchString(nsApp) {
		result.Message = "Invalid app value"
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	labels := map[string]string{}

	// check if environment is allowed
	if ! toolbox.SliceStringContains(app.NamespaceEnvironments, nsEnvironment) {
		result.Message = fmt.Sprintf("Environment \"%s\" not allowed in this cluster", nsEnvironment)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	switch (nsEnvironment) {
	case "team":
		if ! app.RegexpNamespaceTeam.MatchString(nsAreaTeam) {
			result.Message = "Invalid team value"
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}

		if ! c.checkTeamMembership(nsAreaTeam) {
			result.Message = fmt.Sprintf("Access to team \"%s\" denied", nsAreaTeam)
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}

		result.Namespace = fmt.Sprintf("team-%s-%s", nsAreaTeam, nsApp)
		labels["team"] = nsAreaTeam
	case "user":
		result.Namespace = fmt.Sprintf("user-%s-%s", username, nsApp)
		labels["user"] = username
	default:
		if ! c.checkTeamMembership(nsAreaTeam) {
			result.Message = fmt.Sprintf("Access to team \"%s\" denied", nsAreaTeam)
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}

		result.Namespace = fmt.Sprintf("%s-%s", nsEnvironment, nsApp)
		labels["team"] = nsAreaTeam
	}

	namespace := v1.Namespace{}
	namespace.Name = result.Namespace
	namespace.SetLabels(labels)

	if ! c.checkKubernetesNamespaceAccess(namespace) {
		result.Message = fmt.Sprintf("Access to namespace \"%s\" denied", namespace.Name)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	service := services.Kubernetes{}

	// check if already exists
	nsObject, _ := service.NamespaceGet(namespace.Name)
	if nsObject != nil && nsObject.GetUID() != "" {
		result.Message = fmt.Sprintf("Namespace \"%s\" already exists", namespace.Name)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	if err := service.NamespaceCreate(namespace); err != nil {
		result.Message = fmt.Sprintf("Error: %v", err)
		c.Response.Status = http.StatusInternalServerError
	}

	c.Flash.Success("Namespace \"%s\" created", namespace.Name)

	return c.RenderJSON(result)
}

func (c ApiNamespace) Delete(namespace string) revel.Result {
	result := struct {
		Namespace string
		Message string
	} {
		Namespace: namespace,
		Message: "",
	}

	if result.Namespace == "" {
		result.Message = "Invalid namespace"
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	service := services.Kubernetes{}
	nsObject, err := service.NamespaceGet(namespace)

	if err != nil {
		c.Log.Error(fmt.Sprintf("K8S-ERROR: %v", err))
		result.Message = "Communcation error"
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJSON(result)
	}

	if ! c.checkKubernetesNamespaceAccess(*nsObject) {
		result.Message = fmt.Sprintf("Access to namespace \"%s\" denied", result.Namespace)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	if err := service.NamespaceDelete(nsObject.Name); err != nil {
		result.Message = fmt.Sprintf("Error: %v", err)
		c.Response.Status = http.StatusInternalServerError
	}

	c.Flash.Success("Namespace \"%s\" deleted", nsObject.Name)

	return c.RenderJSON(result)
}
