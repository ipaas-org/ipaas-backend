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

type Application struct {
	ID             primitive.ObjectID `bson:"_id" json:"-"`
	ContainerID    string             `bson:"containerID" json:"containerID,omitempty"`
	Status         string             `bson:"status" json:"status,omitempty"`
	StudentID      int                `bson:"studentID" json:"studentID,omitempty"`
	Type           string             `bson:"type" json:"type,omitempty"`
	Name           string             `bson:"name" json:"name,omitempty"`
	Description    string             `bson:"description" json:"description,omitempty"`
	GithubRepo     string             `bson:"githubRepo,omitemtpy" json:"githubRepo,omitempty"`
	GithubBranch   string             `bson:"githubBranch,omitemtpy" json:"githubBranch,omitempty"`
	LastCommitHash string             `bson:"lastCommitHash,omitemtpy" json:"lastCommitHash,omitempty"`
	Port           string             `bson:"port" json:"port,omitempty"`
	ExternalPort   string             `bson:"externalPort" json:"externalPort,omitempty"`
	Lang           string             `bson:"lang" json:"lang,omitempty"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt,omitempty"`
	IsPublic       bool               `bson:"isPublic" json:"isPublic"`
	IsUpdatable    bool               `bson:"isUpdatable,omitempty" json:"isUpdatable"`
	Img            string             `bson:"img,omitempty" json:"img,omitempty"`
	Envs           []Env              `bson:"envs,omitempty" json:"envs,omitempty"`
	Tags           []string           `bson:"tags,omitempty" json:"tags,omitempty"`
	Stars          []string           `bson:"stars,omitempty" json:"stars,omitempty"`
}

type Env struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
}
