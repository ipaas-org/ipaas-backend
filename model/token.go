package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshToken struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserCode   string             `bson:"userCode"`
	Token      string             `bson:"token"`
	Expiration time.Time          `bson:"expiration"`
}
