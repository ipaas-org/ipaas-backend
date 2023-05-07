package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ipaas-org/ipaas-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// get student struct from the id (matricola)
func GetStudentFromID(userID int, connection *mongo.Database) (model.Student, error) {
	var student model.Student
	err := connection.Collection("users").
		FindOne(context.Background(), bson.M{"userID": userID}).
		Decode(&student)
	if err != nil {
		return student, err
	}

	return student, nil
}

// check if the userUID is saved in the db
func IsUserRegistered(userID int, connection *mongo.Database) (bool, error) {
	var user model.Student
	err := connection.Collection("users").
		FindOne(context.Background(), bson.M{"userID": userID}).
		Decode(&user)
	fmt.Println("user found: ", user)
	fmt.Println("student id used:", userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// adds the student to the db (it's a pointer because here we generate the pfp)
func AddStudent(student *model.Student, connection *mongo.Database) error {
	//check if it's already registered
	exists, err := IsUserRegistered(student.ID, connection)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("user already registered")
	}

	student.Pfp = fmt.Sprintf("https://avatars.dicebear.com/api/bottts/%d.svg", student.ID)
	// student.MID = primitive.NewObjectID()
	_, err = connection.Collection("users").InsertOne(context.Background(), student)
	return err
}

// given the paleoid access token and the database connection, it returns the response that
// will be parsed into a json in the response, if the error is client side (4xx or 5xx) and an error
func registerOrGenerateTokenFromPaleoIDAccessToken(paleoidAccess string, connection *mongo.Database) (map[string]interface{}, bool, error) {
	//get the student from the paleoid access token
	student, err := GetStudentFromPaleoIDAccessToken(paleoidAccess)
	if err != nil {
		return nil, true, errors.New("invalid paleoid access token")
	}

	//check if it's in the db
	registered, err := IsUserRegistered(student.ID, connection)
	log.Printf("the user with %d is registered? %t ", student.ID, registered)
	if err != nil {
		return nil, false, err
	}

	//register the user if not already saved
	if !registered {
		log.Println("registering the user")
		err = AddStudent(&student, connection)
		if err != nil {
			return nil, false, err
		}
	}
	//generate the ipaas tokens
	access, refresh, err := GenerateTokenPair(student.ID, connection)
	if err != nil {
		return nil, false, fmt.Errorf("error generating token pair: %v", err)
	}

	//this map will be parsed into a json in the response
	resp := map[string]interface{}{
		"ipaas-access-token":  access,
		"ipaas-refresh-token": refresh,
		"userID":              student.ID,
	}
	return resp, false, nil
}
