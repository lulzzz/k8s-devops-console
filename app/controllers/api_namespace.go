package controllers

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
	"errors"
	"strings"
	"net/http"
	"k8s.io/api/core/v1"
	"github.com/revel/revel"
	"k8s-devops-console/app"
	"k8s-devops-console/app/services"
	"k8s-devops-console/app/toolbox"
	v12 "k8s.io/api/rbac/v1"
	v13 "k8s.io/api/networking/v1"
	"k8s.io/api/settings/v1alpha1"
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
	Labels map[string]string
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

	k8sAnnotationDescription := app.GetConfigString("k8s.annotation.namespace.description", "");

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
			Deleteable: c.checkDeletable(&ns),
		};

		if val, ok := ns.Annotations[k8sAnnotationDescription]; ok {
			row.Description = val
		}

		if val, ok := ns.Labels["team"]; ok {
			row.OwnerTeam = val
		}

		if val, ok := ns.Labels["user"]; ok {
			row.OwnerUser = val
		}

		row.Labels = ns.Labels

		ret = append(ret, row)
	}

	app.PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "listNamespace"}).Inc()

	return c.RenderJSON(ret)
}

func (c ApiNamespace) Create() revel.Result {
	result := struct {
		Namespace string
		Message string
	} {
		Namespace: "",
		Message: "",
	}

	nsEnvironment := ""
	nsAreaTeam := ""
	nsApp := ""
	nsDescription := ""
	nsLabels := map[string]string{}
	c.Params.Bind(&nsEnvironment, "environment")
	c.Params.Bind(&nsAreaTeam, "team")
	c.Params.Bind(&nsApp, "app")
	c.Params.Bind(&nsDescription, "description")
	c.Params.Bind(&nsLabels, "label")

	labelUserKey := app.GetConfigString("k8s.label.user", "user");
	labelTeamKey := app.GetConfigString("k8s.label.team", "team");
	labelEnvKey := app.GetConfigString("k8s.label.environment", "environment");

	user := c.getUser()
	username := user.Username

	if ! app.RegexpNamespaceApp.MatchString(nsApp) {
		result.Message = "Invalid app value"
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	labels := map[string]string{
		labelEnvKey: nsEnvironment,
	}

	// validation
	validationMessages := []string{}
	for _, setting := range app.AppConfig.Kubernetes.Namespace.Labels {
		if val, ok := nsLabels[setting.Name]; ok {
			if setting.Validation.Validate(val) {
				labels[setting.K8sLabel] = val
			} else {
				validationMessages = append(validationMessages, fmt.Sprintf("Validation of \"%s\" failed (%v)", setting.Label, setting.Validation.HumanizeString()))
			}
		}
	}

	if len(validationMessages) >= 1 {
		result.Message = strings.Join(validationMessages, "\n")
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(result)
	}



	// check if environment is allowed
	if ! toolbox.SliceStringContains(app.NamespaceEnvironments, nsEnvironment) {
		result.Message = fmt.Sprintf("Environment \"%s\" not allowed in this cluster", nsEnvironment)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

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

	switch (nsEnvironment) {
	case "team":
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
		labels[labelTeamKey] = strings.ToLower(nsAreaTeam)
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

	k8sAnnotationDescription := app.GetConfigString("k8s.annotation.namespace.description", "");
	if namespace.Annotations == nil {
		namespace.Annotations = map[string]string{}
	}
	namespace.Annotations[k8sAnnotationDescription] = nsDescription

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
		if err := c.updateNamespaceSettings(newNamespace); err != nil {
			result.Message = fmt.Sprintf("%v", err)
			c.Response.Status = http.StatusForbidden
			return c.RenderJSON(result)
		}
	} else {
		result.Message = fmt.Sprintf("Error: %v", err)
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJSON(result)
	}

	c.auditLog(fmt.Sprintf("Namespace \"%s\" created", namespace.Name))
	app.PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "createNamespace"}).Inc()

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

	service := services.Kubernetes{}

	// get namespace
	nsObject, errResult := c.getNamespace(namespace)
	if errResult != nil {
		return *errResult
	}

	if !c.checkDeletable(nsObject) {
		result.Message = fmt.Sprintf("Deletion of namespace \"%s\" denied", result.Namespace)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	if err := service.NamespaceDelete(nsObject.Name); err != nil {
		result.Message = fmt.Sprintf("Error: %v", err)
		c.Response.Status = http.StatusInternalServerError
	}

	c.auditLog(fmt.Sprintf("Namespace \"%s\" deleted", nsObject.Name))
	app.PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "deleteNamepace"}).Inc()

	return c.RenderJSON(result)
}

func (c ApiNamespace) ResetRBAC(namespace string) revel.Result {
	var err error
	result := struct {
		Namespace string
		Message string
	} {
		Namespace: namespace,
		Message: "",
	}

	// get namespace
	nsObject, errResult := c.getNamespace(namespace)
	if errResult != nil {
		return *errResult
	}

	if nsObject, err = c.updateNamespace(nsObject); err != nil {
		result.Message = fmt.Sprintf("%v", err)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	if err := c.updateNamespacePermissions(nsObject); err != nil {
		result.Message = fmt.Sprintf("%v", err)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	result.Message = fmt.Sprintf("Namespace \"%s\" permissions resetted", nsObject.Name)
	c.auditLog(fmt.Sprintf("Namespace \"%s\" permissions resetted", nsObject.Name))
	app.PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "resetRbac"}).Inc()

	return c.RenderJSON(result)
}

func (c ApiNamespace) ResetSettings(namespace string) revel.Result {
	var err error
	result := struct {
		Namespace string
		Message string
	} {
		Namespace: namespace,
		Message: "",
	}

	// get namespace
	nsObject, errResult := c.getNamespace(namespace)
	if errResult != nil {
		return *errResult
	}

	if nsObject, err = c.updateNamespace(nsObject); err != nil {
		result.Message = fmt.Sprintf("%v", err)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	if err := c.updateNamespaceObjects(nsObject); err != nil {
		result.Message = fmt.Sprintf("%v", err)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	result.Message = fmt.Sprintf("Namespace \"%s\" settings resetted", nsObject.Name)
	c.auditLog(fmt.Sprintf("Namespace \"%s\" settings resetted", nsObject.Name))
	app.PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "resetSettings"}).Inc()

	return c.RenderJSON(result)
}

func (c ApiNamespace) updateNamespace(namespace *v1.Namespace) (*v1.Namespace, error) {
	doUpdate := false
	service := services.Kubernetes{}

	labelEnvKey := app.GetConfigString("k8s.label.environment", "environment");

	// add env label
	if _, ok := namespace.Labels[labelEnvKey]; !ok {
		parts := strings.Split(namespace.Name, "-")

		if len(parts) > 1 {
			namespace.Labels[labelEnvKey] = parts[0]
			doUpdate = true
		}
	}

	if doUpdate {
		if _, err := service.NamespaceUpdate(namespace); err != nil {
			return namespace, err
		}
	}

	return namespace, nil
}

func (c ApiNamespace) SetDescription(namespace, description string) revel.Result {
	result := struct {
		Namespace string
		Message string
	} {
		Namespace: namespace,
		Message: "",
	}
	service := services.Kubernetes{}

	// get namespace
	nsObject, errResult := c.getNamespace(namespace)
	if errResult != nil {
		return *errResult
	}

	k8sAnnotationDescription := app.GetConfigString("k8s.annotation.namespace.description", "");

	if nsObject.Annotations == nil {
		nsObject.Annotations = map[string]string{}
	}
	nsObject.Annotations[k8sAnnotationDescription] = description

	if _, err := service.NamespaceUpdate(nsObject); err != nil {
		result.Message = fmt.Sprintf("Access to namespace \"%s\" denied", result.Namespace)
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(result)
	}

	result.Message = fmt.Sprintf("Namespace \"%s\" description changed", nsObject.Name)
	//c.auditLog(fmt.Sprintf("Namespace \"%s\" description changed", nsObject.Name))
	app.PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "setDescription"}).Inc()

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

func (c ApiNamespace) updateNamespaceSettings(namespace *v1.Namespace) (error error) {
	if err := c.updateNamespacePermissions(namespace); err != nil {
		return err
	}

	if err := c.updateNamespaceObjects(namespace); err != nil {
		return err
	}

	return
}

func (c ApiNamespace) updateNamespacePermissions(namespace *v1.Namespace) (error error) {
	service := services.Kubernetes{}

	if !c.checkNamespaceOwnership(namespace) {
		return errors.New(fmt.Sprintf("Namespace \"%s\" not owned by current user", namespace.Name))
	}

	user := c.getUser()
	username := user.Username
	k8sUsername := user.Id

	privateNamespaceEnabled := app.GetConfigBoolean("k8s.user.namespaceRole.private", true)
	labelUserKey := app.GetConfigString("k8s.label.user", "user");
	labelTeamKey := app.GetConfigString("k8s.label.team", "team");

	if labelUserVal, ok := namespace.Labels[labelUserKey]; privateNamespaceEnabled && ok {
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
			for _, permission := range namespaceTeam.K8sPermissions {
				if _, err := service.RoleBindingCreateNamespaceTeam(namespace.Name, labelTeamVal, permission); err != nil {
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

func (c ApiNamespace) updateNamespaceObjects(namespace *v1.Namespace) (error error) {
	var kubeObjectList *app.KubeObjectList
	service := services.Kubernetes{}

	labelEnvKey := app.GetConfigString("k8s.label.environment", "environment");

	if environment, ok := namespace.Labels[labelEnvKey]; ok {
		if configObjects, ok := app.KubeNamespaceConfig[environment]; ok {
			kubeObjectList = configObjects
		}
	}

	// if empty, try default
	if kubeObjectList == nil {
		if configObjects, ok := app.KubeNamespaceConfig["_default"]; ok {
			kubeObjectList = configObjects
		}
	}

	if kubeObjectList != nil {
		for _, kubeObject := range kubeObjectList.ConfigMaps {
			error = service.NamespaceEnsureConfigMap(namespace.Name, kubeObject.Name, kubeObject.Object.(*v1.ConfigMap))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.ServiceAccounts {
			error = service.NamespaceEnsureServiceAccount(namespace.Name, kubeObject.Name, kubeObject.Object.(*v1.ServiceAccount))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.Roles {
			error = service.NamespaceEnsureRole(namespace.Name, kubeObject.Name, kubeObject.Object.(*v12.Role))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.RoleBindings {
			error = service.NamespaceEnsureRoleBindings(namespace.Name, kubeObject.Name, kubeObject.Object.(*v12.RoleBinding))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.NetworkPolicies {
			error = service.NamespaceEnsureNetworkPolicy(namespace.Name, kubeObject.Name, kubeObject.Object.(*v13.NetworkPolicy))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.LimitRanges {
			error = service.NamespaceEnsureLimitRange(namespace.Name, kubeObject.Name, kubeObject.Object.(*v1.LimitRange))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.PodPresets {
			error = service.NamespaceEnsurePodPreset(namespace.Name, kubeObject.Name, kubeObject.Object.(*v1alpha1.PodPreset))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.ResourceQuotas {
			error = service.NamespaceEnsureResourceQuota(namespace.Name, kubeObject.Name, kubeObject.Object.(*v1.ResourceQuota))
			if error != nil {
				return
			}
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

func (c ApiNamespace) checkDeletable(namespace *v1.Namespace) bool {
	ret := app.RegexpNamespaceDeleteFilter.MatchString(namespace.Name)

	annotationImmortal := app.GetConfigString("k8s.annotation.namespace.immortal", "");
	if val, ok := namespace.Annotations[annotationImmortal]; ok {
		if val == "true" {
			ret = false
		}
	}

	if !c.checkNamespaceOwnership(namespace) {
		ret = false
	}

	return ret
}


func (c ApiNamespace) checkNamespaceOwnership(namespace *v1.Namespace) bool {
	user := c.getUser()
	username := user.Username

	labelUserKey := app.GetConfigString("k8s.label.user", "user");
	labelTeamKey := app.GetConfigString("k8s.label.team", "team");

	if labelUserVal, ok := namespace.Labels[labelUserKey]; ok {
		if (labelUserVal == username) {
			return true
		}
	} else if labelTeamVal, ok := namespace.Labels[labelTeamKey]; ok {
		// Team rolebinding
		if _, err := user.GetTeam(labelTeamVal); err == nil {
			return true
		}
	}

	return false
}

func (c ApiNamespace) getNamespace(namespace string) (ns *v1.Namespace, result *revel.Result) {
	resultMessage := struct {
		Namespace string
		Message string
	} {
		Namespace: namespace,
		Message: "",
	}

	if namespace == "" {
		resultMessage.Message = "Invalid namespace"
		c.Response.Status = http.StatusForbidden
		tmp := c.RenderJSON(resultMessage)
		result = &tmp
		return
	}

	service := services.Kubernetes{}
	nsObject, err := service.NamespaceGet(namespace)

	if err != nil {
		c.Log.Error(fmt.Sprintf("K8S-ERROR: %v", err))
		resultMessage.Message = fmt.Sprintf("%s", err)
		c.Response.Status = http.StatusInternalServerError
		tmp := c.RenderJSON(resultMessage)
		result = &tmp
		return
	}

	if ! c.checkKubernetesNamespaceAccess(*nsObject) {
		resultMessage.Message = fmt.Sprintf("Access to namespace \"%s\" denied", namespace)
		c.Response.Status = http.StatusForbidden
		tmp := c.RenderJSON(resultMessage)
		result = &tmp
		return
	}

	ns = nsObject
	return
}
