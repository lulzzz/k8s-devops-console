package models

import (
	"fmt"
	"errors"
	"encoding/json"
)

type User struct {
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

	// Build teams (with permissions)
	for _, teamName := range teamList {
		if _, exists := config.Team[teamName]; exists {
			teamConfig := config.Team[teamName]

			permissions := []TeamPermissions{}

			for _, val := range teamConfig.RoleBinding {
				permissions = append(permissions, TeamPermissions{Name: val.Name, Groups: val.Groups, ClusterRole: val.ClusterRole})
			}

			teams = append(teams, Team{Name: teamName, Permissions: permissions})
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
