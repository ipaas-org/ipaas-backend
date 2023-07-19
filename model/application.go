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
	ServiceType      string
	ApplicationState string

	KeyValue struct {
		Key   string `bson:"key" json:"key"`
		Value string `bson:"value" json:"value"`
	}

	Application struct {
		ID             primitive.ObjectID `bson:"_id,omitemtpy" json:"-"`
		Type           ServiceType        `bson:"type" json:"type,omitempty"`
		Name           string             `bson:"name" json:"name,omitempty"`
		State          ApplicationState   `bson:"state" json:"state,omitempty"`
		OwnerEmail     string             `bson:"ownerEmail" json:"ownerEmail,omitempty"`
		PortToMap      string             `bson:"portToMap" json:"portToMap"`
		Container      Container          `bson:"container" json:"container,omitempty"`
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
	ApplicationTypeWeb      ServiceType = "web"
	ApplicationTypeDatabase ServiceType = "database"

	ApplicationStateCreated ApplicationState = "created"
)

func (a ApplicationState) String() string {
	return string(a)
}
