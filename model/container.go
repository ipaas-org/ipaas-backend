package model

import (
	"time"
)

type Container struct {
	ID              string          `bson:"ID" json:"ID"`
	ImageID         string          `bson:"imageID" json:"imageID"`
	Name            string          `bson:"name" json:"name"`
	Status          ContainerStatus `bson:"status" json:"status"`
	CreatedAt       time.Time       `bson:"createdAt" json:"createdAt"`
	AttachedNetorks []string        `bson:"attachedNetworks,omitempty" json:"attachedNetworks"`
	Labels          []KeyValue      `bson:"labels,omitempty" json:"labels"`
	// Port            string          `bson:"port" json:"port"`
}

type ContainerStatus string

const (
	ContainerCreatedStatus ContainerStatus = "created"
)
