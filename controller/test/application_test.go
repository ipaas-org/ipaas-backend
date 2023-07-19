package controller

import (
	"context"
	"log"
	"testing"

	"github.com/ipaas-org/ipaas-backend/config"
	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/pkg/logger"
	"github.com/ipaas-org/ipaas-backend/repo/mock"
)

func NewController() *controller.Controller {
	conf, err := config.NewConfig("../../config/")
	if err != nil {
		log.Fatalln(err)
	}

	l := logger.NewLogger("debug", "text")
	l.Debug("initizalized logger")
	l.Infof("conf: %+v\n", conf)

	ctx, _ := context.WithCancel(context.Background())
	c := controller.NewController(ctx, conf, l)

	l.Info("using mock database")
	c.SetUserRepo(mock.NewUserRepoer())
	c.SetTokenRepo(mock.NewTokenRepoer())
	c.SetStateRepo(mock.NewStateRepoer())
	c.SetApplicationRepo(mock.NewApplicationRepoer())

	return c
}

// test if the name is available system wide
// [x] check if a name never used is available
// [x] check if a name already used is not available
func TestIsNameAvailableSystemWide(t *testing.T) {
	c := NewController()
	ctx := context.Background()

	name := "test-app"
	if !c.IsNameAvailableSystemWide(ctx, name) {
		t.Errorf("name[%s] should be available", name)
	}
	//this function only check the database
	//so we need to create a new application to test if the name is available
	c.InsertApplication(ctx, &model.Application{
		Name: name,
	})

	if c.IsNameAvailableSystemWide(ctx, name) {
		t.Errorf("name[%s] should not be available", name)
	}
}

func TestIsNameAvailableUserWide(t *testing.T) {
	c := NewController()
	ctx := context.Background()

	user := &model.User{
		Username:  "test-user",
		Role:      model.RoleTesting,
		NetworkID: "test-network-id",
	}
	if err := c.CreateUser(ctx, user); err != nil {
		t.Error(err)
	}

	name := "test-app"
	if !c.IsNameAvailableUserWide(ctx, name, user.Username) {
		t.Errorf("name[%s] should be available", name)
	}

	c.InsertApplication(ctx, &model.Application{
		Name:       name,
		OwnerEmail: user.Username,
	})

	if c.IsNameAvailableUserWide(ctx, name, user.Username) {
		t.Errorf("name[%s] should not be available", name)
	}
}

// func TestCreateNewContainer(t *testing.T) {
// 	c := NewController()
// 	ctx := context.Background()

// 	username := "test-user"
// 	appname := "test-app"
// 	imageID := "4c408489c41a"
// 	id, containerName, err := c.CreateNewContainer(ctx, controller.WebType, username, appname, imageID, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Logf("id: %s, containerName: %s", id, containerName)

// 	c.StartContainer(ctx, id)
// }
