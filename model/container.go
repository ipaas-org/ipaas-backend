package model

import (
	"time"
)

type Container struct {
	ContainerID     string          `bson:"containerID" json:"containerID"`
	ImageID         string          `bson:"imageID" json:"imageID"`
	Name            string          `bson:"name" json:"name"`
	Status          ContainerStatus `bson:"status" json:"status"`
	CreatedAt       time.Time       `bson:"createdAt" json:"createdAt"`
	Envs            []KeyValue      `bson:"envs,omitempty" json:"envs"`
	Labels          []KeyValue      `bson:"labels,omitempty" json:"labels"`
	AttachedNetorks []string        `bson:"attachedNetworks,omitempty" json:"attachedNetworks"`
	// Port            string          `bson:"port" json:"port"`
}

type ContainerStatus string

const (
	ContainerCreatedStatus ContainerStatus = "created"
)
