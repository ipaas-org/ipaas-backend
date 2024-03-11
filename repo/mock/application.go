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

func (r *ApplicationRepoerMock) FindByID(ctx context.Context, _id primitive.ObjectID) (*model.Application, error) {
	entity, ok := r.storage[_id]
	if ok {
		return entity, nil
	}
	return nil, repo.ErrNotFound
}

func (r *ApplicationRepoerMock) FindByName(ctx context.Context, name string) (*model.Application, error) {
	for _, entity := range r.storage {
		if entity.Name == name {
			return entity, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (r *ApplicationRepoerMock) FindByNameAndOwner(ctx context.Context, name, owner string) (*model.Application, error) {
	for _, entity := range r.storage {
		if entity.Name == name && entity.Owner == owner {
			return entity, nil
		}
	}
	return nil, repo.ErrNotFound
}

// func (r *ApplicationRepoerMock) FindByContainerID(ctx context.Context, containerID string) (*model.Application, error) {
// 	for _, entity := range r.storage {
// 		if entity.Implementation != nil && entity.Implementation.ID == containerID {
// 			return entity, nil
// 		}
// 	}
// 	return nil, repo.ErrNotFound
// }

func (r *ApplicationRepoerMock) FindByOwner(ctx context.Context, owner string) ([]*model.Application, error) {
	var entities []*model.Application
	for _, entity := range r.storage {
		if entity.Owner == owner {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (r *ApplicationRepoerMock) FindByOwnerAndKind(ctx context.Context, owner string, kind model.ApplicationKind) ([]*model.Application, error) {
	var entities []*model.Application
	for _, entity := range r.storage {
		if entity.Owner == owner && entity.Kind == kind {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (r *ApplicationRepoerMock) FindByOwnerAndKindAndIsPublicTrue(ctx context.Context, owner string, serviceType model.ApplicationKind) ([]*model.Application, error) {
	var entities []*model.Application
	for _, entity := range r.storage {
		if entity.Owner == owner && entity.Kind == serviceType && entity.IsPublic {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (r *ApplicationRepoerMock) FindByOwnerAndKindAndIsPublicFalse(ctx context.Context, owner string, serviceType model.ApplicationKind) ([]*model.Application, error) {
	var entities []*model.Application
	for _, entity := range r.storage {
		if entity.Owner == owner && entity.Kind == serviceType && !entity.IsPublic {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (r *ApplicationRepoerMock) FindByOwnerAndIsUpdatableTrue(ctx context.Context, owner string) ([]*model.Application, error) {
	var entities []*model.Application
	for _, entity := range r.storage {
		if entity.Owner == owner && entity.IsUpdatable {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (r *ApplicationRepoerMock) InsertOne(ctx context.Context, application *model.Application) (interface{}, error) {
	id := primitive.NewObjectID()
	if application.ID != primitive.NilObjectID {
		id = application.ID
	}
	application.ID = id
	r.storage[id] = application
	return id, nil
}

func (r *ApplicationRepoerMock) UpdateByID(ctx context.Context, application *model.Application, _id primitive.ObjectID) (bool, error) {
	_, ok := r.storage[_id]
	if !ok {
		return false, repo.ErrNotFound
	}
	r.storage[_id] = application
	return true, nil
}

func (r *ApplicationRepoerMock) DeleteByID(ctx context.Context, _id primitive.ObjectID) (bool, error) {
	_, ok := r.storage[_id]
	if !ok {
		return false, repo.ErrNotFound
	}
	delete(r.storage, _id)
	return true, nil
}
