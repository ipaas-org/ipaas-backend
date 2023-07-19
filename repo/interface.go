package repo

import (
	"context"
	"errors"

	"github.com/ipaas-org/ipaas-backend/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	UserRepoer interface {
		InsertOne(ctx context.Context, userModel *model.User) (interface{}, error)
		FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
		FindByEmail(ctx context.Context, email string) (*model.User, error)
		UpdateGithubAccessTokenByID(ctx context.Context, githubAccessToken string, id primitive.ObjectID) (bool, error)
		DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error)
	}

	TokenRepoer interface {
		InsertOne(ctx context.Context, t *model.RefreshToken) (id interface{}, err error)
		FindByToken(ctx context.Context, token string) (*model.RefreshToken, error)
		DeleteByToken(ctx context.Context, token string) (bool, error)
	}

	StateRepoer interface {
		InsertOne(ctx context.Context, s *model.State) (id interface{}, err error)
		FindByState(ctx context.Context, state string) (*model.State, error)
		DeleteByState(ctx context.Context, state string) (bool, error)
	}

	ApplicationRepoer interface {
		InsertOne(ctx context.Context, a *model.Application) (id interface{}, err error)
		FindByID(ctx context.Context, id primitive.ObjectID) (*model.Application, error)
		FindByName(ctx context.Context, name string) (*model.Application, error)
		FindByNameAndOwnerUsername(ctx context.Context, name, ownerUsername string) (*model.Application, error)
		FindByContainerID(ctx context.Context, containerID string) (*model.Application, error)
		FindByOwnerUsername(ctx context.Context, ownerUsername string) ([]*model.Application, error)
		FindByOwnerUsernameAndTypeAndIsPublicTrue(ctx context.Context, ownerUsername string, appType model.ServiceType) ([]*model.Application, error)
		FindByOwnerUsernameAndTypeAndIsPublicFalse(ctx context.Context, ownerUsername string, appType model.ServiceType) ([]*model.Application, error)
		FindByOwnerUsernameAndIsUpdatableTrue(ctx context.Context, ownerUsername string) ([]*model.Application, error)
		UpdateByID(ctx context.Context, a *model.Application, id primitive.ObjectID) (bool, error)
		DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error)
	}
)

var (
	ErrNotFound = errors.New("not found")
)
