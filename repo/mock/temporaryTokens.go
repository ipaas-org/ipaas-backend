package mock

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
)

type content struct {
	access  *model.AccessToken
	refresh *model.RefreshToken
}

func NewTemporaryTokenRepoer() repo.TemporaryTokenStorage {
	return &TempTokenRepoerMock{
		storage: make(map[string]*content),
	}
}

type TempTokenRepoerMock struct {
	storage map[string]*content
}

func (r *TempTokenRepoerMock) FindByKey(ctx context.Context, key string) (*model.AccessToken, *model.RefreshToken, error) {
	entity, ok := r.storage[key]
	if !ok {
		return nil, nil, repo.ErrNotFound
	}
	return entity.access, entity.refresh, nil
}

func (r *TempTokenRepoerMock) InsertTokens(ctx context.Context, key string, access *model.AccessToken, refresh *model.RefreshToken) error {
	c := new(content)
	c.access = access
	c.refresh = refresh
	r.storage[key] = c
	return nil
}

func (r *TempTokenRepoerMock) DeleteKey(ctx context.Context, key string) error {
	_, ok := r.storage[key]
	if !ok {
		return repo.ErrNotFound
	}
	delete(r.storage, key)
	return nil
}
