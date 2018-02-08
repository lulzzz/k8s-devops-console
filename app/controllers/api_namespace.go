package controllers

import (
	"fmt"
	"regexp"
	"errors"
	"strings"
	"net/http"
	"k8s.io/api/core/v1"
	"github.com/revel/revel"
	"k8s-devops-console/app"
	"k8s-devops-console/app/services"
	"k8s-devops-console/app/toolbox"
)

type ResultNamespace struct {
	Name string
	Environment string
	Description string
	OwnerTeam string
	OwnerUser string
	Status string
	Created string
	CreatedAgo string
	Deleteable bool
}

type ApiNamespace struct {
	ApiBase
}

func (c ApiNamespace) accessCheck() (result revel.Result) {
	return c.ApiBase.accessCheck()
}

func (c ApiNamespace) List() revel.Result {
	service := services.Kubernetes{}
	nsList, err := service.NamespaceList()
	if err != nil {
		c.Log.Error(fmt.Sprintf("K8s communication error: %v", err))
		return c.renderJSONError("Unable to contact cluster")
	}

	ret := []ResultNamespace{}

	for _, ns := range nsList {
		if ! c.checkKubernetesNamespaceAccess(ns) {
			continue;
		}

		namespaceParts := strings.Split(ns.Name, "-")
		environment := ""
		if len(namespaceParts) > 2 {
			environment = namespaceParts[0]
		}

		row := ResultNamespace{
			Name: ns.Name,
			Environment: environment,
			Status: fmt.Sprintf("%v", ns.Status.Phase),
			Created: ns.CreationTimestamp.UTC().String(),
			CreatedAgo: revel.TimeAgo(ns.CreationTimestamp.UTC()),
			Deleteable: app.RegexpNamespaceDeleteFilter.MatchString(ns.Name),
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

	labelUserKey := app.GetConfigString("k8s.label.user", "user");
	labelTeamKey := app.GetConfigString("k8s.label.team", "team");

	roleBinding := "team"
	user := c.getUser()
	username := user.Username
	k8sUsername := user.Id

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
		// team filter check
		if !app.RegexpNamespaceTeam.MatchString(nsAreaTeam)  {
			result.Message = "Invalid team value"
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}

		// membership check
		if ! c.checkTeamMembership(nsAreaTeam) {
			result.Message = fmt.Sprintf("Access to team \"%s\" denied", nsAreaTeam)
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}

		// quota check
		if err := c.checkNamespaceTeamQuota(nsAreaTeam); err != nil {
			result.Message = fmt.Sprintf("Error: %v", err)
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}

		result.Namespace = fmt.Sprintf("team-%s-%s", nsAreaTeam, nsApp)
		labels[labelTeamKey] = strings.ToLower(nsAreaTeam)
	case "user":
		// quota check
		if err := c.checkNamespaceUserQuota(username); err != nil {
			result.Message = fmt.Sprintf("Error: %v", err)
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}

		result.Namespace = fmt.Sprintf("user-%s-%s", username, nsApp)
		labels[labelUserKey] = strings.ToLower(username)
		roleBinding = "user"
	default:
		// membership check
		if !c.checkTeamMembership(nsAreaTeam) {
			result.Message = fmt.Sprintf("Access to team \"%s\" denied", nsAreaTeam)
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}

		result.Namespace = fmt.Sprintf("%s-%s", nsEnvironment, nsApp)
		labels[labelTeamKey] = strings.ToLower(nsAreaTeam)
	}

	// filtering
	result.Namespace = strings.ToLower(result.Namespace)
	result.Namespace = strings.Replace(result.Namespace, "_", "", -1)

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
	existingNs, _ := service.NamespaceGet(namespace.Name)
	if existingNs != nil && existingNs.GetUID() != "" {
		if existingNsTeam, ok := existingNs.Labels["team"]; ok {
			result.Message = fmt.Sprintf("Namespace \"%s\" already exists (owned by team \"%s\")", namespace.Name, existingNsTeam)
		} else if existingNsUser, ok := existingNs.Labels["user"]; ok {
			result.Message = fmt.Sprintf("Namespace \"%s\" already exists (owned by user \"%s\")", namespace.Name, existingNsUser)
		} else {
			result.Message = fmt.Sprintf("Namespace \"%s\" already exists", namespace.Name)
		}

		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	// Namespace creation
	if newNamespace, err := service.NamespaceCreate(namespace); newNamespace != nil && err == nil {
		if err := c.setNamespacePermissions(newNamespace); err != nil {
			result.Message = fmt.Sprintf("%v", err)
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}
	} else {
		result.Message = fmt.Sprintf("Error: %v", err)
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJSON(result)
	}

	switch roleBinding {
	case "team":
		// Team rolebinding
		if namespaceTeam, err := user.GetTeam(nsAreaTeam); err == nil {
			for _, permission := range namespaceTeam.Permissions {
				if _, err := service.RoleBindingCreateNamespaceTeam(namespace.Name, nsAreaTeam, permission.Name, permission.Groups, permission.ClusterRole); err != nil {
					result.Message = fmt.Sprintf("Error: %v", err)
					c.Response.Status = http.StatusInternalServerError
				}
			}
		}
	case "user":
		// User rolebinding
		role := app.GetConfigString("k8s.user.namespaceRole", "admin")
		if _, err := service.RoleBindingCreateNamespaceUser(namespace.Name, username, k8sUsername, role); err != nil {
			result.Message = fmt.Sprintf("Error: %v", err)
			c.Response.Status = http.StatusInternalServerError
		}
	}

	// ServiceAccount rolebinding
	if role := app.GetConfigString("k8s.serviceaccount.namespaceRole", ""); role != "" {
		if _, err := service.RoleBindingCreateNamespaceServiceAccount(namespace.Name, "default", role); err != nil {
			result.Message = fmt.Sprintf("Error: %v", err)
			c.Response.Status = http.StatusInternalServerError
		}
	}


	c.auditLog(fmt.Sprintf("Namespace \"%s\" created", namespace.Name))

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
		result.Message = fmt.Sprintf("%s", err)
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJSON(result)
	}

	if ! c.checkKubernetesNamespaceAccess(*nsObject) {
		result.Message = fmt.Sprintf("Access to namespace \"%s\" denied", result.Namespace)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	if !app.RegexpNamespaceDeleteFilter.MatchString(namespace) {
		result.Message = fmt.Sprintf("Deletion of namespace \"%s\" denied", result.Namespace)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	if err := service.NamespaceDelete(nsObject.Name); err != nil {
		result.Message = fmt.Sprintf("Error: %v", err)
		c.Response.Status = http.StatusInternalServerError
	}

	c.auditLog(fmt.Sprintf("Namespace \"%s\" deleted", nsObject.Name))

	return c.RenderJSON(result)
}

func (c ApiNamespace) ResetPermissions(namespace string) revel.Result {
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
		result.Message = fmt.Sprintf("%s", err)
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJSON(result)
	}

	if ! c.checkKubernetesNamespaceAccess(*nsObject) {
		result.Message = fmt.Sprintf("Access to namespace \"%s\" denied", result.Namespace)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	if err := c.setNamespacePermissions(nsObject); err != nil {
		result.Message = fmt.Sprintf("%v", err)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	result.Message = fmt.Sprintf("Namespace \"%s\" permissions resetted", nsObject.Name)
	c.auditLog(fmt.Sprintf("Namespace \"%s\" permissions resetted", nsObject.Name))

	return c.RenderJSON(result)
}

func (c ApiNamespace) checkNamespaceTeamQuota(team string) (err error) {
	var count int
	quota := app.GetConfigInt("k8s.namespace.team.quota", 0)

	if quota <= 0 {
		// no quota
		return
	}

	regexp := regexp.MustCompile(fmt.Sprintf(app.NamespaceFilterTeam, regexp.QuoteMeta(team)));

	service := services.Kubernetes{}
	count, err = service.NamespaceCount(regexp)
	if err != nil {
		return
	}

	if count >= quota {
		// quota exceeded
		err = errors.New(fmt.Sprintf("Team namespace quota of %v namespaces exceeded ", quota))
	}

	return
}

func (c ApiNamespace) setNamespacePermissions(namespace *v1.Namespace) (error error) {
	service := services.Kubernetes{}

	user := c.getUser()
	username := user.Username
	k8sUsername := user.Id

	labelUserKey := app.GetConfigString("k8s.label.user", "user");
	labelTeamKey := app.GetConfigString("k8s.label.team", "team");

	if labelUserVal, ok := namespace.Labels[labelUserKey]; ok {
		if (labelUserVal == username) {
			// User rolebinding
			role := app.GetConfigString("k8s.user.namespaceRole", "admin")
			if _, err := service.RoleBindingCreateNamespaceUser(namespace.Name, username, k8sUsername, role); err != nil {
				return errors.New(fmt.Sprintf("Error: %v", err))
			}
		} else {
			return errors.New(fmt.Sprintf("Namespace \"%s\" not owned by current user", namespace.Name))
		}
	} else if labelTeamVal, ok := namespace.Labels[labelTeamKey]; ok {
		// Team rolebinding
		if namespaceTeam, err := user.GetTeam(labelTeamVal); err == nil {
			for _, permission := range namespaceTeam.Permissions {
				if _, err := service.RoleBindingCreateNamespaceTeam(namespace.Name, labelTeamVal, permission.Name, permission.Groups, permission.ClusterRole); err != nil {
					return errors.New(fmt.Sprintf("Error: %v", err))
				}
			}
		}
	} else {
		return errors.New(fmt.Sprintf("Namespace \"%s\" cannot be resetted, labels not found", namespace.Name))
	}

	// ServiceAccount rolebinding
	if role := app.GetConfigString("k8s.serviceaccount.namespaceRole", ""); role != "" {
		if _, err := service.RoleBindingCreateNamespaceServiceAccount(namespace.Name, "default", role); err != nil {
			return errors.New(fmt.Sprintf("Error: %v", err))
		}
	}

	return
}

func (c ApiNamespace) checkNamespaceUserQuota(username string) (err error) {
	var count int
	quota := app.GetConfigInt("k8s.namespace.user.quota", 0)

	if quota <= 0 {
		// no quota
		return
	}

	regexp := regexp.MustCompile(fmt.Sprintf(app.NamespaceFilterUser, regexp.QuoteMeta(username)));

	service := services.Kubernetes{}
	count, err = service.NamespaceCount(regexp)
	if err != nil {
		return
	}

	if count >= quota {
		// quota exceeded
		err = errors.New(fmt.Sprintf("Personal namespace quota of %v namespaces exceeded ", quota))
	}

	return
}
