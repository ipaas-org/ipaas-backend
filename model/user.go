package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID            primitive.ObjectID `bson:"_id" json:"_"`
	StudentNumber int                `bson:"studentID" json:"matricola"` //id of the student (teachers will use the same, but it's 6 digit long instead of 5)
	Name          string             `bson:"name" json:"nome"`
	LastName      string             `bson:"lastName" json:"cognome"`
	Email         string             `bson:"email" json:"email"`
	Pfp           string             `bson:"pfp" json:"pfp"`
}

// !currently unused
type StudentInfo struct {
	Class               string `json:"classe"`
	Year                int    `json:"anno"`
	Field               string `json:"indirizzo"`
	IsClassPresident    bool   `json:"rappresentante_classe"`
	IsIstiturePresident bool   `json:"rappresentante_istituto"`
}
