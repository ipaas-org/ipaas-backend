package model

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	RoleUser    Role = "user"
	RoleTesting Role = "testing" //used only for unit testing
)

type (
	User struct {
		ID           primitive.ObjectID `bson:"_id,omitemtpy" json:"-"`
		Code         string             `bson:"code" json:"code"`
		NetworkID    string             `bson:"networkId" json:"networkId"`
		Role         Role               `bson:"role" json:"role"` //defaults to "user"
		UserSettings *UserSettings      `bson:"userSettings" json:"userSettings"`
		Info         *UserInfo          `bson:"userInfo" json:"userInfo"`
	}

	UserInfo struct {
		Username          string `bson:"username" json:"username"`   //github oauth field: login
		FullName          string `bson:"name" json:"name"`           //github oauth field: name
		Email             string `bson:"email" json:"email"`         //github oauth field: email primary
		Pfp               string `bson:"pfp" json:"pfp"`             //github oauth field: avatar_url
		GithubUrl         string `bson:"githubUrl" json:"githubUrl"` //github oauth field: url
		GithubAccessToken string `bson:"githubAccessToken" json:"-"`
	}

	UserSettings struct {
		Theme string `bson:"theme" json:"theme"`
	}

	Role string
)

func (u *User) String() string {
	// return fmt.Sprintf("%s [ID=%s]", u.Username, u.ID.Hex())
	return fmt.Sprintf("UID=%s [_ID=%s](username=%q name=%q email=%q)", u.Code, u.ID.Hex(), u.Info.Username, u.Info.FullName, u.Info.Email)
}
