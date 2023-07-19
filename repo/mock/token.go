package mock

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewTokenRepoer() repo.TokenRepoer {
	return &TokenRepoerMock{
		storage: make(map[string]*model.RefreshToken),
	}
}

type TokenRepoerMock struct {
	storage map[string]*model.RefreshToken
}

func (r *TokenRepoerMock) FindByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	entity, ok := r.storage[token]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return entity, nil
}

func (r *TokenRepoerMock) InsertOne(ctx context.Context, token *model.RefreshToken) (interface{}, error) {
	id := primitive.NewObjectID()
	if token.ID != primitive.NilObjectID {
		id = token.ID
	}
	token.ID = id
	r.storage[token.Token] = token
	return id, nil
}

func (r *TokenRepoerMock) DeleteByToken(ctx context.Context, token string) (bool, error) {
	_, ok := r.storage[token]
	if !ok {
		return false, repo.ErrNotFound
	}
	delete(r.storage, token)
	return true, nil
}
