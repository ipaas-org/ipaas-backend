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
		ID            primitive.ObjectID `bson:"_id,omitempty" json:"-"`
		Kind          ServiceKind        `bson:"kind" json:"kind"`
		Name          string             `bson:"name" json:"name"`
		DnsName       string             `bson:"dnsName" json:"dnsName"`
		State         ApplicationState   `bson:"state" json:"state"`
		Owner         string             `bson:"owner" json:"owner"`
		CreatedAt     time.Time          `bson:"createdAt" json:"createdAt,omitempty"`
		ListeningPort string             `bson:"listeningPort" json:"listeningPort"`
		Description   string             `bson:"description,omitempty" json:"description,omitempty"`
		GithubRepo    string             `bson:"githubRepo" json:"githubRepo"`
		GithubBranch  string             `bson:"githubBranch" json:"githubBranch"`
		BuiltCommit   string             `bson:"builtCommit" json:"builtCommit,omitempty"`
		IsPublic      bool               `bson:"isPublic" json:"isPublic"`
		IsUpdatable   bool               `bson:"isUpdatable,omitempty" json:"isUpdatable"`
		Container     *Container         `bson:"container" json:"container,omitempty"`
		Envs          []KeyValue         `bson:"envs,omitempty" json:"envs"`
		// Image          *Image             `bson:"image" json:"image,omitempty"`
	}
)

const (
	ApplicationKindWeb      ServiceKind = "web"
	ApplicationKindDatabase ServiceKind = "database"

	ApplicationStatePending  ApplicationState = "pending"
	ApplicationStateBuilding ApplicationState = "building"
	ApplicationStateStarting ApplicationState = "starting"
	ApplicationStateRunning  ApplicationState = "running"
	ApplicationStateFailed   ApplicationState = "failed"
)

func (s ServiceKind) String() string {
	return string(s)
}

func (a ApplicationState) String() string {
	return string(a)
}
