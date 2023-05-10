package mock

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewStateRepoer() repo.StateRepoer {
	return &StateRepoerMock{
		storage: make(map[string]*model.State),
	}
}

type StateRepoerMock struct {
	storage map[string]*model.State
}

func (r *StateRepoerMock) FindByState(ctx context.Context, state string) (*model.State, error) {
	entity, ok := r.storage[state]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return entity, nil
}

func (r *StateRepoerMock) Insert(ctx context.Context, state *model.State) (interface{}, error) {
	id := primitive.NewObjectID()
	state.ID = id
	r.storage[state.State] = state
	return id, nil
}

func (r *StateRepoerMock) DeleteByState(ctx context.Context, state string) (bool, error) {
	_, ok := r.storage[state]
	if !ok {
		return false, repo.ErrNotFound
	}
	delete(r.storage, state)
	return true, nil
}
