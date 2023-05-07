package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// struct used to get the user from paleoid
type (
	Payload struct {
		GrantType    string `json:"grant_type"`    //will always be "authorization_code"
		Code         string `json:"code"`          //the code returned by the oauth server
		RedirectUri  string `json:"redirect_uri"`  //the redirect uri (saved in env variable)
		ClientID     string `json:"client_id"`     //the client id (saved in env variable)
		ClientSecret string `json:"client_secret"` //the client secret (saved in env variable)
	}

	State struct {
		Id             primitive.ObjectID `bson:"_id"`
		State          string             `json:"state"`
		Issued         time.Time          `bson:"issDate"`
		ExpirationDate time.Time          `bson:"expDate"`
		RedirectUri    string             `bson:"redirectUri"`
	}

	Polling struct {
		DBId            primitive.ObjectID `bson:"_id"`
		RandomId        string             `bson:"id"`
		IssDate         time.Time          `bson:"issDate"`
		ExpDate         time.Time          `bson:"expDate"`
		LoginSuccessful bool               `bson:"loginSuccessful"`
	}
)

const (
	GrantTypeAuthorizationCode = "authorization_code"
	BaseUrlPaleoID                 = "https://id.paleo.bg.it/"
)
