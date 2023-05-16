package mock

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewApplicationRepoer() repo.ApplicationRepoer {
	return &ApplicationRepoerMock{
		storage: make(map[primitive.ObjectID]*model.Application),
	}
}

type ApplicationRepoerMock struct {
	storage map[primitive.ObjectID]*model.Application
}

func (r *ApplicationRepoerMock) FindByID(arg0 context.Context, arg1 primitive.ObjectID) (*model.Application, error) {
	entity, ok := r.storage[arg1]
	if ok {
		return entity, nil
	}
	return nil, repo.ErrNotFound
}

func (r *ApplicationRepoerMock) FindByName(arg0 context.Context, arg1 string) (*model.Application, error) {
	for _, entity := range r.storage {
		if entity.Name == arg1 {
			return entity, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (r *ApplicationRepoerMock) FindByNameAndOwnerUsername(ctx context.Context, name, ownerUsername string) (*model.Application, error) {
	for _, entity := range r.storage {
		if entity.Name == name && entity.OwnerUsername == ownerUsername {
			return entity, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (r *ApplicationRepoerMock) FindByContainerID(arg0 context.Context, arg1 string) (*model.Application, error) {
	for _, entity := range r.storage {
		if entity.ContainerID == arg1 {
			return entity, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (r *ApplicationRepoerMock) FindByOwnerUsername(arg0 context.Context, arg1 string) ([]*model.Application, error) {
	var entities []*model.Application
	for _, entity := range r.storage {
		if entity.OwnerUsername == arg1 {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (r *ApplicationRepoerMock) FindByOwnerUsernameAndTypeAndIsPublicTrue(arg0 context.Context, arg1 string, arg2 string) ([]*model.Application, error) {
	var entities []*model.Application
	for _, entity := range r.storage {
		if entity.OwnerUsername == arg1 && entity.Type == arg2 && entity.IsPublic {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (r *ApplicationRepoerMock) FindByOwnerUsernameAndTypeAndIsPublicFalse(arg0 context.Context, arg1 string, arg2 string) ([]*model.Application, error) {
	var entities []*model.Application
	for _, entity := range r.storage {
		if entity.OwnerUsername == arg1 && entity.Type == arg2 && !entity.IsPublic {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (r *ApplicationRepoerMock) FindByOwnerUsernameAndIsUpdatableTrue(arg0 context.Context, arg1 string) ([]*model.Application, error) {
	var entities []*model.Application
	for _, entity := range r.storage {
		if entity.OwnerUsername == arg1 && entity.IsUpdatable {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (r *ApplicationRepoerMock) Insert(arg0 context.Context, arg1 *model.Application) (interface{}, error) {
	id := primitive.NewObjectID()
	arg1.ID = id
	r.storage[id] = arg1
	return id, nil
}

func (r *ApplicationRepoerMock) UpdateByID(arg0 context.Context, arg1 *model.Application, arg2 primitive.ObjectID) (bool, error) {
	_, ok := r.storage[arg2]
	if !ok {
		return false, repo.ErrNotFound
	}
	r.storage[arg2] = arg1
	return true, nil
}

func (r *ApplicationRepoerMock) DeleteByID(arg0 context.Context, arg1 primitive.ObjectID) (bool, error) {
	_, ok := r.storage[arg1]
	if !ok {
		return false, repo.ErrNotFound
	}
	delete(r.storage, arg1)
	return true, nil
}
