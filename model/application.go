package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	ApplicationKind  string
	ApplicationState string

	KeyValue struct {
		Key   string `bson:"key" json:"key"`
		Value string `bson:"value" json:"value"`
	}

	Application struct {
		ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
		CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
		UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
		Name          string             `bson:"name" json:"name"`
		Kind          ApplicationKind    `bson:"kind" json:"kind"`
		DnsName       string             `bson:"dnsName" json:"dnsName"`
		State         ApplicationState   `bson:"state" json:"state"`
		Owner         string             `bson:"owner" json:"owner"`
		ListeningPort string             `bson:"listeningPort" json:"listeningPort"`
		Description   string             `bson:"description,omitempty" json:"description,omitempty"`
		GithubRepo    string             `bson:"githubRepo" json:"githubRepo"`
		GithubBranch  string             `bson:"githubBranch" json:"githubBranch"`
		BuiltCommit   string             `bson:"builtCommit" json:"builtCommit,omitempty"`
		Visiblity     string             `bson:"visiblity" json:"visiblity"`
		IsUpdatable   bool               `bson:"isUpdatable" json:"isUpdatable"`
		Service       *Service           `bson:"service" json:"-"`
		Envs          []KeyValue         `bson:"envs" json:"envs"`
		BasedOn       string             `bson:"basedOn" json:"basedOn"` //id of the template the application is based on
		// Image          *Image             `bson:"image" json:"image,omitempty"`
	}
)

const (
	ApplicationKindWeb       ApplicationKind = "web"
	ApplicationKindStorage   ApplicationKind = "storage"
	ApplicationKindManagment ApplicationKind = "managment"

	ApplicationStatePending  ApplicationState = "pending"
	ApplicationStateBuilding ApplicationState = "building"
	ApplicationStateStarting ApplicationState = "starting"
	ApplicationStateRunning  ApplicationState = "running"
	ApplicationStateFailed   ApplicationState = "failed"
	ApplicationStateDeleting ApplicationState = "deleting"
	ApplicationStateCrashed  ApplicationState = "crashed"

	ApplicationVisiblityPublic  = "public"
	ApplicationVisiblityPrivate = "private"
)

func (s ApplicationKind) String() string {
	return string(s)
}

func (a ApplicationState) String() string {
	return string(a)
}
