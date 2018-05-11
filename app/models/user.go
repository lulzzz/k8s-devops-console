package models

import (
	"fmt"
	"sort"
	"errors"
	"encoding/json"
)

type User struct {
	Uuid     string `json:"Uuid"`
	Id       string `json:"id"`
	Username string `json:"u"`
	Email    string `json:"e"`
	Teams    []Team `json:"-"`
	Groups   []string `json:"g"`
	IsAdmin  bool `json:"a"`
}

func (u *User) init(config *AppConfig) {
	u.initTeams(config)
}

func (u *User) initTeams(config *AppConfig) (teams []Team) {
	teamList := map[string]string{}

	if config == nil {
		return
	}

	// Default teams
	for _, teamName := range config.Default.Teams {
		teamList[teamName] = teamName
	}

	// User teams
	if config.User != nil {
		if val, exists := config.User[u.Username]; exists {
			for _, teamName := range val.Teams {
				teamList[teamName] = teamName
			}
		}
	}

	// Group teams
	if config.Group != nil {
		for _, group := range u.Groups {
			if val, exists := config.Group[group]; exists {
				for _, teamName := range val.Teams {
					teamList[teamName] = teamName
				}
			}
		}
	}

	// Sort
	teamNameList := make([]string, 0, len(teamList))
	for teamName := range teamList {
		teamNameList = append(teamNameList, teamName)
	}
	sort.Strings(teamNameList)

	// Build teams (with permissions)
	for _, teamName := range teamNameList {
		if _, exists := config.Team[teamName]; exists {
			teamConfig := config.Team[teamName]
			teams = append(teams, Team{Name: teamName, K8sPermissions: teamConfig.K8sRoleBinding, AzureRoleAssignments: teamConfig.AzureRoleAssignments})
		}
	}

	u.Teams = teams
	return
}

func (u *User) GetTeam(name string) (team *Team, err error) {
	for _, val := range u.Teams {
		if val.Name == name {
			team = &val
			break
		}
	}

	if team == nil {
		err = errors.New("Team not found")
	}

	return
}

func (u *User) ToJson() (jsonString string, error error) {
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		error = err
		return
	}

	jsonString = string(jsonBytes)
	return
}

func UserCreateFromJson(jsonString string, config *AppConfig) (u *User, err error) {
	if err = json.Unmarshal([]byte(jsonString), &u); err == nil {
		u.init(config)
	}

	return
}

func (u *User) String() string {
	return fmt.Sprintf("User(%s)", u.Username)
}
