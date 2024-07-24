package controller

import (
	"context"
	"github.com/ipaas-org/ipaas-backend/handlers/httpserver"
	"log"
	"testing"
	"time"

	"github.com/ipaas-org/ipaas-backend/config"
	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/pkg/logger"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/ipaas-org/ipaas-backend/repo/mock"
	mongoRepo "github.com/ipaas-org/ipaas-backend/repo/mongo"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// test if the name is available system-wide
func TestIsNameAvailableSystemWide(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	// check if an unused app-name is available in the system
	t.Run("Check unused app-name", func(t *testing.T) {
		name := "test-app"
		if !c.IsNameAvailableSystemWide(ctx, name) {
			t.Errorf("app name [%v] should be available", name)
		}

		if err := c.InsertApplication(ctx, &model.Application{
			Name: name,
		}); err != nil {
			t.Errorf("error inserting app: %v", err)
		}
	})

	// check if a used app-name isn't available in the system
	t.Run("Check used app-name", func(t *testing.T) {
		name := "test-app"
		if !c.IsNameAvailableSystemWide(ctx, name) {
			t.Errorf("app name [%v] should be available", name)
		}

		if err := c.InsertApplication(ctx, &model.Application{
			Name: name,
		}); err != nil {
			t.Errorf("error inserting app: %v", err)
		}

		if c.IsNameAvailableSystemWide(ctx, name) {
			t.Errorf("app name [%v] should not be available", name)
		}
	})

	cancel()
}

// test if an application result existing in the system
func TestDoesApplicationExists(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	// check if the unused app-id refers to an application
	t.Run("Check unused app-id", func(t *testing.T) {
		id := primitive.NewObjectID()
		if exist, err := c.DoesApplicationExists(ctx, id); err != nil {
			t.Errorf("error checking app: %v", err)
		} else if exist {
			t.Errorf("app id [%v] should not exist", id)
		}
	})

	// check if the used app-id refers to an application
	t.Run("Check used app-id", func(t *testing.T) {
		id := primitive.NewObjectID()
		if exist, err := c.DoesApplicationExists(ctx, id); err != nil {
			t.Errorf("error checking app: %v", err)
		} else if exist {
			t.Errorf("app id [%v] should not exist", id)
		}

		name := "test-app"
		if err := c.InsertApplication(ctx, &model.Application{
			Name: name,
		}); err != nil {
			t.Errorf("error inserting: %v", err)
		}

		app, err := c.ApplicationRepo.FindByName(ctx, name)
		if err != nil {
			t.Errorf("error finding: %v", err)
		}

		if exist, err := c.DoesApplicationExists(ctx, app.ID); err != nil {
			t.Errorf("error checking: %v", err)
		} else if !exist {
			t.Errorf("app id [%v] should exist", id)
		}
	})

	cancel()
}

// test if an application can be retrieved by its own app-id
func TestGetApplicationByID(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	// check if the unused app-id refers to an application
	t.Run("Check unused app-id", func(t *testing.T) {
		id := primitive.NewObjectID()
		_, err := c.GetApplicationByID(ctx, id)
		if err == nil {
			t.Errorf("app id [%v] should not exist", id)
		}
	})

	// check if the used app-id refers to an application
	t.Run("Check used app-id", func(t *testing.T) {
		name := "test-app"
		if err := c.InsertApplication(ctx, &model.Application{
			Name: name,
		}); err != nil {
			t.Errorf("error inserting app: %v", err)
		}

		app, err := c.ApplicationRepo.FindByName(ctx, name)
		if err != nil {
			t.Errorf("error finding app: %v", err)
		}

		_, err = c.GetApplicationByID(ctx, app.ID)
		if err != nil {
			t.Errorf("app id [%v] should exist", app.ID)
		}
	})

	cancel()
}

// test if all application of one user can be retrieved
func TestGetAllUserApplications(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	// check for a user without applications
	t.Run("Check user with 0 app", func(t *testing.T) {
		userInfo := &model.UserInfo{
			Username: "test-user",
		}
		user, err := c.CreateUser(ctx, userInfo, model.RoleTesting)
		if err != nil {
			t.Errorf("error creating user: %v", err)
		}

		apps, err := c.GetAllUserApplications(ctx, user.Code)
		if err != nil {
			t.Errorf("error getting user apps: %v", err)
		}

		if len(apps) != 0 {
			t.Errorf("user [%v] should not have any apps", user.Code)
		}
	})

	// check for a user with 1 application
	t.Run("Check user with 1 app", func(t *testing.T) {
		userInfo := &model.UserInfo{
			Username: "test-user",
		}
		user, err := c.CreateUser(ctx, userInfo, model.RoleTesting)
		if err != nil {
			t.Errorf("error creating user: %v", err)
		}

		name := "test-app"
		if err := c.InsertApplication(ctx, &model.Application{
			Name:  name,
			Owner: user.Code,
		}); err != nil {
			t.Errorf("error inserting app: %v", err)
		}

		apps, err := c.GetAllUserApplications(ctx, user.Code)
		if err != nil {
			t.Errorf("error getting user apps: %v", err)
		}

		app, err := c.ApplicationRepo.FindByName(ctx, name)
		if err != nil {
			t.Errorf("error finding app: %v", err)
		}

		if len(apps) != 1 {
			t.Errorf("user [%v] should have one app", user.Code)
		} else if apps[0].ID != app.ID {
			t.Errorf("app id [%v] and app id [%v] should be equal", apps[0].ID, app.ID)
		}
	})

	cancel()
}

// test if all applications of one user can be retrieved filtered by its own kind
func TestGetApplicationByKind(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	// check for a user without applications
	t.Run("Check user with 0 app", func(t *testing.T) {
		userInfo := &model.UserInfo{
			Username: "test-user",
		}
		user, err := c.CreateUser(ctx, userInfo, model.RoleTesting)
		if err != nil {
			t.Errorf("error creating user: %v", err)
		}

		apps, err := c.GetApplicationByKind(ctx, user.Code, model.ApplicationKindWeb)
		if err != nil {
			t.Errorf("error getting user apps: %v", err)
		}

		if len(apps) != 0 {
			t.Errorf("user [%v] should not have any apps", user.Code)
		}
	})

	// check for a user with 1 application of a different kind that the one searched for
	t.Run("Check user with 1 app", func(t *testing.T) {
		userInfo := &model.UserInfo{
			Username: "test-user",
		}
		user, err := c.CreateUser(ctx, userInfo, model.RoleTesting)
		if err != nil {
			t.Errorf("error creating user: %v", err)
		}

		name := "test-app"
		if err := c.InsertApplication(ctx, &model.Application{
			Name:  name,
			Owner: user.Code,
			Kind:  model.ApplicationKindManagment,
		}); err != nil {
			t.Errorf("error inserting app: %v", err)
		}

		apps, err := c.GetApplicationByKind(ctx, user.Code, model.ApplicationKindWeb)
		if err != nil {
			t.Errorf("error getting user apps: %v", err)
		}

		if len(apps) != 0 {
			t.Errorf("user [%v] should not have any apps of 'Web' kind", user.Code)
		}
	})

	// check for a user with 1 application of the same kind as the one searched for
	t.Run("Check user with 1 app", func(t *testing.T) {
		userInfo := &model.UserInfo{
			Username: "test-user",
		}
		user, err := c.CreateUser(ctx, userInfo, model.RoleTesting)
		if err != nil {
			t.Errorf("error creating user: %v", err)
		}

		name := "test-app"
		if err := c.InsertApplication(ctx, &model.Application{
			Name:  name,
			Owner: user.Code,
			Kind:  model.ApplicationKindWeb,
		}); err != nil {
			t.Errorf("error inserting app: %v", err)
		}

		apps, err := c.GetApplicationByKind(ctx, user.Code, model.ApplicationKindWeb)
		if err != nil {
			t.Errorf("error getting user apps: %v", err)
		}

		app, err := c.ApplicationRepo.FindByName(ctx, name)
		if err != nil {
			t.Errorf("error finding app: %v", err)
		}

		if len(apps) != 1 {
			t.Errorf("user id [%v] should have 1 application", user.Code)
		} else if apps[0].ID != app.ID {
			t.Errorf("app id [%v] and app id [%v] should be equal", apps[0].ID, app.ID)
		}
	})

	cancel()
}

// test if an application can be inserted in the system
func TestInsertApplication(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	// check if a new unique application can be inserted
	t.Run("Check insertion of new app", func(t *testing.T) {
		name := "test-app"
		_, err := c.ApplicationRepo.FindByName(ctx, name)
		if err != nil && err != repo.ErrNotFound {
			t.Errorf("error finding app: %v", err)
		} else if err == nil {
			t.Errorf("app name [%v] should not exist", name)
		}

		if err := c.InsertApplication(ctx, &model.Application{
			Name: name,
		}); err != nil {
			t.Errorf("error inserting app: %v", err)
		}

		_, err = c.ApplicationRepo.FindByName(ctx, name)
		if err != nil {
			if err != repo.ErrNotFound {
				t.Errorf("error finding app: %v", err)
			} else {
				t.Errorf("app name [%v] should exist", name)
			}
		}
	})

	// check if an already inserted application can't be inserted anymore
	t.Run("Check insertion of already inserted app", func(t *testing.T) {
		name := "test-app"
		app, err := c.ApplicationRepo.FindByName(ctx, name)
		if err != nil && err != repo.ErrNotFound {
			t.Errorf("error finding app: %v", err)
		} else if err == nil {
			t.Errorf("app [%v] should not exist", app.ID)
		}

		if err := c.InsertApplication(ctx, &model.Application{
			Name: name,
		}); err != nil {
			t.Errorf("error inserting app: %v", err)
		}

		app, err = c.ApplicationRepo.FindByName(ctx, name)
		if err != nil {
			if err != repo.ErrNotFound {
				t.Errorf("error finding app: %v", err)
			} else {
				t.Errorf("app [%v] should exist", app.ID)
			}
		}

		if err := c.InsertApplication(ctx, &model.Application{
			Name: name,
		}); err == nil {
			t.Errorf("app [%v] shouldn't be inserted", app.ID)
		}
	})

	cancel()
}

// TODO: obtain user github infos and app build config in order to test a WebApp creation
// test if a web application can be created, initialized and its relative image built
func TestCreateNewWebApplication(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	// check for a basic valid image
	t.Run("Check creation of minimal test app", func(t *testing.T) {
		userInfo := &model.UserInfo{
			Username: "test-user",
		}
		user, err := c.CreateUser(ctx, userInfo, model.RoleTesting)
		if err != nil {
			t.Errorf("error creating user: %v", err)
		}

		post := &httpserver.HttpWebApplicationPost{
			Name:        "test-app",
			Repo:        "Vano2903/testing",
			Branch:      "master",
			Port:        "8080",
			Description: "test-app description",
			Envs: []model.KeyValue{
				{Key: "Key", Value: "Value"},
			},
			BuildConfig: &model.BuildConfig{},
		}
		if _, err := c.CreateNewWebApplication(ctx, user.Code, user.Info.GithubAccessToken, post.Name, post.Repo, post.Branch, post.Port, post.Envs, post.BuildConfig); err != nil {
			t.Errorf("error creating app: %v", err)
		}

		app, err := c.ApplicationRepo.FindByName(ctx, post.Name)
		if err != nil {
			t.Error("error finding app: %v", err)
		} else if app.Owner != user.Code {
			t.Errorf("app owner [%v] should be user code [%v]", app.Owner, user.Code)
		}
	})

	cancel()
}

// TODO: obtain image infos in order to test an App (from id) creation
func TestCreateApplicationFromApplicationIDandImageID(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	// check for creation from invalid application id
	t.Run("check with invalid app id", func(t *testing.T) {
		//TODO: implement this unit test
	})

	// check for creation from invalid image id
	t.Run("check with invalid image id", func(t *testing.T) {
		//TODO: implement this unit test
	})

	// check for creation from valid both application id and image id
	t.Run("check with valid app id and image id", func(t *testing.T) {
		userInfo := &model.UserInfo{
			Username: "test-user",
		}
		user, err := c.CreateUser(ctx, userInfo, model.RoleTesting)
		if err != nil {
			t.Errorf("error creating user: %v", err)
		}

		name := "test-app"
		if err := c.InsertApplication(ctx, &model.Application{
			Name:          name,
			Owner:         user.Code,
			Kind:          model.ApplicationKindWeb,
			ListeningPort: "8080",
			Description:   "test-app description",
			GithubRepo:    "Vano2903/testing",
			GithubBranch:  "master",
			Envs: []model.KeyValue{
				{Key: "Key", Value: "Value"},
			},
		}); err != nil {
			t.Errorf("error inserting app: %v", err)
		}

		app, err := c.ApplicationRepo.FindByName(ctx, name)
		if err != nil {
			t.Errorf("error finding app: %v", err)
		}

		//TODO:
		imageID := "4c408489c41a"
		buildCommit := "4c408489c41a"

		if err := c.CreateApplicationFromApplicationIDandImageID(ctx, app.ID.String(), imageID, buildCommit); err != nil {
			t.Errorf("error creating app: %v", err)
		}
	})

	cancel()
}

// test if an application can be deleted
func TestDeleteApplication(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	// check for an existing application
	t.Run("", func(t *testing.T) {
		userInfo := &model.UserInfo{
			Username: "test-user",
		}
		user, err := c.CreateUser(ctx, userInfo, model.RoleTesting)
		if err != nil {
			t.Errorf("error creating user: %v", err)
		}

		name := "test-app"
		if err := c.InsertApplication(ctx, &model.Application{
			Name:  name,
			Owner: user.Code,
		}); err != nil {
			t.Errorf("error inserting app: %v", err)
		}

		app, err := c.ApplicationRepo.FindByName(ctx, name)
		if err != nil {
			t.Errorf("error finding app: %v", err)
		}

		if err := c.DeleteApplication(ctx, app, user); err != nil {
			t.Errorf("error deleting app: %v", err)
		}

		if _, err := c.ApplicationRepo.FindByName(ctx, name); err != repo.ErrNotFound {
			t.Errorf("app name [%v] should not exist", name)
		}
	})

	cancel()
}

//TODO: Obtain build response infos in order to force an App build fail
// func TestFailedBuild(t *testing.T) {
// 	c, cancel := NewController()
// 	ctx := context.Background()

// 	buildResponse := &model.BuildResponse{}
// 	if err := c.FailedBuild(ctx, buildResponse); err != nil {
// 		t.Errorf("error failing build: %v", err)
// 	}
// 	cancel()
// }

func TestRedeployApplication(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	userInfo := &model.UserInfo{
		Username: "test-user",
	}
	user, err := c.CreateUser(ctx, userInfo, model.RoleTesting)
	if err != nil {
		t.Errorf("error creating user: %v", err)
	}

	name := "test-app"
	if err := c.InsertApplication(ctx, &model.Application{
		Name:          name,
		Owner:         user.Code,
		ListeningPort: "8080",
	}); err != nil {
		t.Errorf("error inserting app: %v", err)
	}

	app, err := c.ApplicationRepo.FindByName(ctx, name)
	if err != nil {
		t.Errorf("error finding app: %v", err)
	}

	if err := c.RedeployApplication(ctx, user, app); err != nil {
		t.Errorf("error redeploying app: %v", err)
	}
	cancel()
}

func TestAddCurrentPodToApplication(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	userInfo := &model.UserInfo{
		Username: "test-user",
	}
	user, err := c.CreateUser(ctx, userInfo, model.RoleTesting)
	if err != nil {
		t.Errorf("error creating user: %v", err)
	}

	name := "test-app"
	if err := c.InsertApplication(ctx, &model.Application{
		Name:          name,
		Owner:         user.Code,
		ListeningPort: "8080",
	}); err != nil {
		t.Errorf("error inserting app: %v", err)
	}

	app, err := c.ApplicationRepo.FindByName(ctx, name)
	if err != nil {
		t.Errorf("error finding app: %v", err)
	}

	podname := "test-podname"
	c.AddCurrentPodToApplication(ctx, app.ID, podname)

	if app.Service.Deployment.CurrentPodName != podname {
		t.Errorf("app pod name [%v] should be pod name [%v]", app.Service.Deployment.CurrentPodName, podname)
	}

	cancel()

}

func TestUpdateApplication(t *testing.T) {
	c, cancel := NewController()
	ctx := context.Background()

	userInfo := &model.UserInfo{
		Username: "test-user",
	}
	user, err := c.CreateUser(ctx, userInfo, model.RoleTesting)
	if err != nil {
		t.Errorf("error creating user: %v", err)
	}

	name := "test-app"
	if err := c.InsertApplication(ctx, &model.Application{
		Name:          name,
		Owner:         user.Code,
		ListeningPort: "8080",
	}); err != nil {
		t.Errorf("error inserting app: %v", err)
	}

	app, err := c.ApplicationRepo.FindByName(ctx, name)
	if err != nil {
		t.Errorf("error finding app: %v", err)
	}

	patchPort := "8081"
	if err := c.UpdateApplication(ctx, app, user, app.Name, patchPort, app.Envs); err != nil {
		t.Errorf("error updating app: %v", err)
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
