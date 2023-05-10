package mock

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewUserRepoer() repo.UserRepoer {
	return &UserRepoerMock{
		storage: make(map[primitive.ObjectID]*model.User),
	}
}

type UserRepoerMock struct {
	storage map[primitive.ObjectID]*model.User
}

func (r *UserRepoerMock) InsertOne(ctx context.Context, user *model.User) (interface{}, error) {
	id := primitive.NewObjectID()
	user.ID = id
	r.storage[id] = user
	return id, nil
}

func (r *UserRepoerMock) UpdateGithubAccessTokenByID(ctx context.Context, githubAccessToken string, id primitive.ObjectID) (bool, error) {
	entity, ok := r.storage[id]
	if !ok {
		return false, repo.ErrNotFound
	}
	entity.GithubAccessToken = githubAccessToken
	return true, nil
}

func (r *UserRepoerMock) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	entity, ok := r.storage[id]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return entity, nil
}

func (r *UserRepoerMock) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	for _, entity := range r.storage {
		if entity.Email == email {
			return entity, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (r *UserRepoerMock) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	for _, entity := range r.storage {
		if entity.Username == username {
			return entity, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (r *UserRepoerMock) DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error) {
	_, ok := r.storage[id]
	if !ok {
		return false, repo.ErrNotFound
	}
	delete(r.storage, id)
	return true, nil
}
