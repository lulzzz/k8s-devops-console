package controllers

import (
	"os"
	"fmt"
	"time"
	"context"
	"k8s-devops-console/app/models"
	"k8s-devops-console/app"
	"github.com/revel/revel"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/hashicorp/go-uuid"
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


func (c ApiAzure) CreateResourceGroup(resourceGroupName, location, team string, personal bool) revel.Result {
	var err error
	var group resources.Group

	ret := true
	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
	user := c.getUser()

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
	group, _ = groupsClient.Get(ctx, resourceGroupName)
	if group.ID != nil {
		tagList := []string{}

		for tagName, tagValue := range group.Tags {
			tagList = append(tagList, fmt.Sprintf("%v=%v", tagName, to.String(tagValue)))
		}

		tagLine := ""
		if len(tagList) >= 1 {
			tagLine = fmt.Sprintf(" tags:%v", tagList)
		}

		c.Log.Error(fmt.Sprintf("Azure ResourceGroup already exists: \"%s\"%s", resourceGroupName, tagLine))
		return c.renderJSONError(fmt.Sprintf("Azure ResourceGroup already exists: \"%s\"%s", resourceGroupName, tagLine))
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


	// create resourceGroup
	tags := map[string]*string{
		"creator": to.StringPtr(user.Username),
		"owner": to.StringPtr(team),
		"updated": to.StringPtr(time.Now().Local().Format("2006-01-02")),
		"created-by": to.StringPtr("devops-console"),
	}
	resourceGroup := resources.Group{
		Location: to.StringPtr(location),
	  	Tags: tags,
	}

	group, err = groupsClient.CreateOrUpdate(ctx, resourceGroupName, resourceGroup)
	if err != nil {
		c.Log.Error(fmt.Sprintf("Unable to create Azure ResourceGroup: %v", err))
		return c.renderJSONError("Unable to create Azure ResourceGroup")
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

	return c.RenderJSON(ret)
}


