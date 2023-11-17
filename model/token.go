package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserCode  string             `bson:"userCode"`
	Token     string             `bson:"token"`
	ExpiresAt time.Time          `bson:"expiresAt"`
}

type AccessToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}
