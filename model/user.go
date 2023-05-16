package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type (
	User struct {
		ID                primitive.ObjectID `bson:"_id" json:"-"`
		Username          string             `bson:"username" json:"username"`   //github oauth field: login
		FullName          string             `bson:"name" json:"name"`           //github oauth field: name
		Email             string             `bson:"email" json:"email"`         //github oauth field: email primary
		Pfp               string             `bson:"pfp" json:"pfp"`             //github oauth field: avatar_url
		GithubUrl         string             `bson:"githubUrl" json:"githubUrl"` //github oauth field: url
		GithubAccessToken string             `bson:"githubAccessToken" json:"-"`
		NetworkID         string             `bson:"networkId" json:"networkId"`
		Role              string             `bson:"role" json:"role"` //defaults to "user"
		UserSettings      UserSettings       `bson:"userSettings" json:"userSettings"`
	}

	UserSettings struct {
		Theme string `bson:"theme" json:"theme"`
	}
)
