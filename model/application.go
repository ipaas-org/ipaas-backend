package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DbContainerConfig struct {
	Name  string
	Image string
	Port  string
}

type (
	ServiceKind      string
	ApplicationState string

	KeyValue struct {
		Key   string `bson:"key" json:"key"`
		Value string `bson:"value" json:"value"`
	}

	Application struct {
		ID             primitive.ObjectID `bson:"_id,omitemtpy" json:"-"`
		Kind           ServiceKind        `bson:"kind" json:"kind,omitempty"`
		Name           string             `bson:"name" json:"name,omitempty"`
		State          ApplicationState   `bson:"state" json:"state,omitempty"`
		Owner          string             `bson:"owner" json:"owner,omitempty"`
		PortToMap      string             `bson:"portToMap" json:"portToMap"`
		Container      *Container         `bson:"container" json:"container,omitempty"`
		Envs           []KeyValue         `bson:"envs,omitempty" json:"envs"`
		Description    string             `bson:"description,omitemtpy" json:"description,omitempty"`
		GithubRepo     string             `bson:"githubRepo,omitemtpy" json:"githubRepo,omitempty"`
		GithubBranch   string             `bson:"githubBranch,omitemtpy" json:"githubBranch,omitempty"`
		LastCommitHash string             `bson:"lastCommitHash,omitemtpy" json:"lastCommitHash,omitempty"`
		CreatedAt      time.Time          `bson:"createdAt" json:"createdAt,omitempty"`
		IsPublic       bool               `bson:"isPublic" json:"isPublic"`
		IsUpdatable    bool               `bson:"isUpdatable,omitempty" json:"isUpdatable"`
	}
)

const (
	ApplicationKindWeb      ServiceKind = "web"
	ApplicationKindDatabase ServiceKind = "database"

	ApplicationStatePending  ApplicationState = "pending"
	ApplicationStateBuilding ApplicationState = "building"
)

func (s ServiceKind) String() string {
	return string(s)
}

func (a ApplicationState) String() string {
	return string(a)
}
