package controller

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/ipaas-org/ipaas-backend/config"
	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/pkg/logger"
	"github.com/ipaas-org/ipaas-backend/repo/mock"
	mongoRepo "github.com/ipaas-org/ipaas-backend/repo/mongo"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectToRepository(conf *config.Config, l *logrus.Logger, c *controller.Controller) {
	switch conf.Database.Driver {
	case "mock":
		l.Info("using mock database")
		c.UserRepo = mock.NewUserRepoer()
		c.TokenRepo = mock.NewTokenRepoer()
		c.StateRepo = mock.NewStateRepoer()
		c.ApplicationRepo = mock.NewApplicationRepoer()

	case "mongo":
		l.Info("using mongo database")

		l.Debug("connecting to database")
		ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.Database.URI))
		if err != nil {
			l.Fatalf("main - mongo.Connect - error connecting to database: %s", err.Error())
		}
		cancel()

		l.Debug("connecting to user collection")
		userCollection := client.Database("ipaas").Collection("user")
		userRepo := mongoRepo.NewUserRepoer(userCollection)
		c.UserRepo = userRepo

		l.Debug("connecting to token collection")
		tokenCollection := client.Database("ipaas").Collection("token")
		tokenRepo := mongoRepo.NewTokenRepoer(tokenCollection)
		c.TokenRepo = tokenRepo

		l.Debug("connecting to state collection")
		stateCollection := client.Database("ipaas").Collection("state")
		stateRepo := mongoRepo.NewStateRepoer(stateCollection)
		c.StateRepo = stateRepo

		l.Debug("connecting to application collection")
		applicationCollection := client.Database("ipaas").Collection("application")
		applicationRepo := mongoRepo.NewApplicationRepoer(applicationCollection)
		c.ApplicationRepo = applicationRepo
	default:
		l.Fatalf("main - unknown database driver: %s", conf.Database.Driver)
	}
}

func NewController() (*controller.Controller, context.CancelFunc) {
	conf, err := config.NewConfig("../../")
	if err != nil {
		log.Fatalln(err)
	}

	l := logger.NewLogger("debug", "text")
	l.Debug("initizalized logger")
	l.Infof("conf: %+v\n", conf)

	ctx, cancel := context.WithCancel(context.Background())
	c := controller.NewController(ctx, conf, l)

	conf.Database.Driver = "mock"
	ConnectToRepository(conf, l, c)

	return c, cancel
}

// test if the name is available system wide
// [x] check if a name never used is available
// [x] check if a name already used is not available
func TestIsNameAvailableSystemWide(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	name := "test-app"
	if !c.IsNameAvailableSystemWide(ctx, name) {
		t.Errorf("name[%s] should be available", name)
	}
	//this function only check the database
	//so we need to create a new application to test if the name is available
	if err := c.InsertApplication(ctx, &model.Application{
		Name: name,
	}); err != nil {
		t.Errorf("error inserting: %v", err)
	}

	if c.IsNameAvailableSystemWide(ctx, name) {
		t.Errorf("name[%s] should not be available", name)
	}
	cancel()
}

func TestIsNameAvailableUserWide(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	info := &model.UserInfo{
		Username: "test-user",
	}
	user, err := c.CreateUser(ctx, info, model.RoleTesting)
	if err != nil {
		t.Error(err)
	}

	name := "test-app"
	if !c.IsNameAvailableUserWide(ctx, name, user.Code) {
		t.Errorf("name[%s] should be available", name)
	}

	if err := c.InsertApplication(ctx, &model.Application{
		Name:  name,
		Owner: user.Code,
	}); err != nil {
		t.Errorf("error inserting: %v", err)
	}

	if c.IsNameAvailableUserWide(ctx, name, user.Code) {
		t.Errorf("name[%s] should not be available", name)
	}
	cancel()
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
