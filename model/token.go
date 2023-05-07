package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshToken struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserID     int                `bson:"userID"`
	Token      string             `bson:"token"`
	Expiration time.Time          `bson:"expiration"`
}
