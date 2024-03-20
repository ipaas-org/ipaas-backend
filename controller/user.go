package controller

import (
	"context"

	"github.com/google/uuid"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *Controller) DoesUserExist(ctx context.Context, email string) (bool, error) {
	_, err := c.UserRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == repo.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Controller) createNewUserCode(ctx context.Context) (string, error) {
	ran, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return "us-" + ran.String(), nil
}

// todo: create user should generate the user not ask for a user model
// todo: separate create user to generate user model (it can ask for the user info but it needs to create the user code and network id,...)
// todo: create insert user function that adds to repo
func (c *Controller) CreateUser(ctx context.Context, info *model.UserInfo, role model.Role) (*model.User, error) {
	user := new(model.User)
	user.Info = info

	userCode, err := c.createNewUserCode(ctx)
	if err != nil {
		return nil, err
	}
	user.Code = userCode

	namespace := "ns-" + userCode
	// networkID, err = c.serviceManager.CreateNewNetwork(ctx, userCode)
	labels := c.getDefaultLabels(userCode, staticTempEnvironment, "", "", namespace)
	if err := c.ServiceManager.CreateNewNamespace(ctx, namespace, labels); err != nil {
		c.l.Errorf("error creating new network: %v", err)
		return nil, err
	}
	user.Namespace = namespace

	if _, err := c.ServiceManager.CreateNewRegistrySecret(ctx, namespace, c.config.K8s.RegistryUrl, c.config.K8s.RegistryUsername, c.config.K8s.RegistryPassword); err != nil {
		c.l.Errorf("error creating new registry secret: %v", err)
		return nil, err
	}

	if role == "" {
		role = model.RoleUser
	}

	user.Role = role
	user.UserSettings = &model.UserSettings{Theme: "light"}

	if err := c.insertUser(ctx, user); err != nil {
		c.l.Errorf("error inserting user %v", user)
		return nil, err
	}
	c.l.Infof("user %v created successfully", user)
	return user, nil
}

func (c *Controller) insertUser(ctx context.Context, user *model.User) error {
	user.ID = primitive.NewObjectID()
	_, err := c.UserRepo.InsertOne(ctx, user)
	return err
}

func (c *Controller) GetUserFromEmail(ctx context.Context, email string) (*model.User, error) {
	return c.UserRepo.FindByEmail(ctx, email)
}

// TODO
func (c *Controller) UpdateUser(ctx context.Context, user *model.User) error {
	return nil
}

// todo: delete user, stop and delete all containers, delete all volumes
// delete network, remove associated tokens
func (c *Controller) DeleteUser(ctx context.Context, email string) (bool, error) {
	// user, err := c.userRepo.FindByEmail(ctx, email)
	// if err != nil {
	// 	return false, err
	// }

	// deleted, err := c.userRepo.DeleteByID(ctx, user.ID)
	// if err != nil {
	// 	return false, err
	// }
	// c.RemoveNetwork(ctx, user.NetworkID)
	return false, nil
}
