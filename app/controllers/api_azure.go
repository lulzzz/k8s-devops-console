package controllers

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/go-uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/revel/revel"
	"k8s-devops-console/app"
	"k8s-devops-console/app/models"
	"os"
	"strings"
	"time"
)

type ResultAzureResourceGroup struct {
	Name string
	OwnerTeam string
	OwnerUser string
	Created string
	CreatedAgo string
	Deleteable bool
}

var (
	ctx        = context.Background()
	authorizer autorest.Authorizer
)

type ApiAzure struct {
	ApiBase
}

func (c ApiAzure) accessCheck() (result revel.Result) {
	return c.ApiBase.accessCheck()
}


func (c ApiAzure) CreateResourceGroup(name, location, team string, personal bool, tag map[string]string) revel.Result {
	var err error
	var group resources.Group

	validationMessages := []string{}

	ret := true
	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
	user := c.getUser()

	// validate name
	if !app.AppConfig.Azure.ResourceGroup.Validation.Validate(name) {
		validationMessages = append(validationMessages,fmt.Sprintf("Validation of ResourceGroup name failed (%v)", app.AppConfig.Azure.ResourceGroup.Validation.HumanizeString()))
	}

	roleAssignmentList := []models.TeamAzureRoleAssignments{}
	roleAssignmentList = append(roleAssignmentList, models.TeamAzureRoleAssignments{
		Role: "Owner",
		PrincipalId: user.Uuid,
	})

	// team filter check
	if !app.RegexpNamespaceTeam.MatchString(team)  {
		c.Log.Error(fmt.Sprintf("Invalid team value", err))
		return c.renderJSONError(fmt.Sprintf("Invalid team value", team))
	}

	// membership check
	if ! c.checkTeamMembership(team) {
		c.Log.Error(fmt.Sprintf("Access to team \"%s\" denied", err))
		return c.renderJSONError(fmt.Sprintf("Access to team \"%s\" denied", team))
	}

	if teamObj, err := user.GetTeam(team); err == nil {
		for _, teamRoleAssignment := range teamObj.AzureRoleAssignments {
			if (personal) {
				teamRoleAssignment.Role = "Reader"
			}

			roleAssignmentList = append(roleAssignmentList,teamRoleAssignment)
		}
	}

	// create ResourceGroup tagList
	tagList := map[string]*string{}

	// add tags from user
	for _, tagConfig := range app.AppConfig.Azure.ResourceGroup.Tags {
		tagValue := ""
		if val, ok := tag[tagConfig.Name]; ok {
			tagValue = val
		}

		if ! tagConfig.Validation.Validate(tagValue) {
			validationMessages = append(validationMessages, fmt.Sprintf("Validation of \"%s\" failed (%v)", tagConfig.Label, tagConfig.Validation.HumanizeString() ))
		}

		if tagValue != "" {
			tagList[tagConfig.Name] = to.StringPtr(tagValue)
		}
	}

	// fixed tags
	tagList["creator"] = to.StringPtr(user.Username)
	tagList["owner"] = to.StringPtr(team)
	tagList["updated"] = to.StringPtr(time.Now().Local().Format("2006-01-02"))
	tagList["created-by"] = to.StringPtr("devops-console")

	if len(validationMessages) >= 1 {
		messages := strings.Join(validationMessages, "\n")
		c.Log.Error(messages)
		return c.renderJSONError(messages)
	}

	// azure authorizer
	authorizer, err = auth.NewAuthorizerFromEnvironment()
	if err != nil {
		c.Log.Error(fmt.Sprintf("Unable to setup Azure Authorizer: %v", err))
		return c.renderJSONError("Unable to setup Azure Authorizer")
	}

	// setup clients
	groupsClient := resources.NewGroupsClient(subscriptionId)
	groupsClient.Authorizer = authorizer

	roleDefinitionsClient := authorization.NewRoleDefinitionsClient(subscriptionId)
	roleDefinitionsClient.Authorizer = authorizer

	roleAssignmentsClient := authorization.NewRoleAssignmentsClient(subscriptionId)
	roleAssignmentsClient.Authorizer = authorizer

	// check for existing resourcegroup
	group, _ = groupsClient.Get(ctx, name)
	if group.ID != nil {
		tagList := []string{}

		for tagName, tagValue := range group.Tags {
			tagList = append(tagList, fmt.Sprintf("%v=%v", tagName, to.String(tagValue)))
		}

		tagLine := ""
		if len(tagList) >= 1 {
			tagLine = fmt.Sprintf(" tags:%v", tagList)
		}

		c.Log.Error(fmt.Sprintf("Azure ResourceGroup already exists: \"%s\"%s", name, tagLine))
		return c.renderJSONError(fmt.Sprintf("Azure ResourceGroup already exists: \"%s\"%s", name, tagLine))
	}

	// translate roles
	var roleAssignmentId string
	for roleAssignmentKey, roleAssignment := range roleAssignmentList {
		// get role definition
		filter := fmt.Sprintf("roleName eq '%s'", roleAssignment.Role)
		roleDefinitions, err := roleDefinitionsClient.List(ctx, "", filter)

		if len(roleDefinitions.Values()) != 1 {
			c.Log.Error(fmt.Sprintf("Error generating UUID for Role Assignment: %v", err))
			return c.renderJSONError(fmt.Sprintf("Error generating UUID for Role Assignment: %v", err))
		}


		// create uuid
		roleAssignmentId, err = uuid.GenerateUUID()
		if err != nil {
			c.Log.Error(fmt.Sprintf("Unable to build UUID: %v", err))
			return c.renderJSONError("Unable to build UUID")
		}

		roleAssignmentList[roleAssignmentKey].Uuid = roleAssignmentId
		roleAssignmentList[roleAssignmentKey].Role = *roleDefinitions.Values()[0].ID
	}

	resourceGroup := resources.Group{
		Location: to.StringPtr(location),
	  	Tags: tagList,
	}

	group, err = groupsClient.CreateOrUpdate(ctx, name, resourceGroup)
	if err != nil {
		c.Log.Error(fmt.Sprintf("Unable to create Azure ResourceGroup: %v", err))
		return c.renderJSONError("Unable to create Azure ResourceGroup")
	}

	if personal {
		c.auditLog(fmt.Sprintf("Azure ResourceGroup \"%s\" created (personal access)", name))

	} else {
		c.auditLog(fmt.Sprintf("Azure ResourceGroup \"%s\" created (team access)", name))
	}

	// assign role to ResourceGroup
	for _, roleAssignment := range roleAssignmentList {
		// assign role to ResourceGroup
		properties := authorization.RoleAssignmentCreateParameters{
			Properties: &authorization.RoleAssignmentProperties{
				RoleDefinitionID: &roleAssignment.Role,
				PrincipalID:      &roleAssignment.PrincipalId,
			},
		}

		_, err = roleAssignmentsClient.Create(ctx, to.String(group.ID), roleAssignment.Uuid, properties)
		if err != nil {
			c.Log.Error(fmt.Sprintf("Unable to create Azure RoleAssignment: %v", err))
			return c.renderJSONError("Unable to create Azure RoleAssignment")
		}
	}

	app.PrometheusActions.With(prometheus.Labels{"scope": "azure", "type": "createResourceGroup"}).Inc()

	return c.RenderJSON(ret)
}
