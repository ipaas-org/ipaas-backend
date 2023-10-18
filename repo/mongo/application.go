package mongo

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewApplicationRepoer(collection *mongo.Collection) repo.ApplicationRepoer {
	return &ApplicationRepoerMongo{
		collection: collection,
	}
}

type ApplicationRepoerMongo struct {
	collection *mongo.Collection
}

func (r *ApplicationRepoerMongo) FindByID(ctx context.Context, _id primitive.ObjectID) (*model.Application, error) {
	var entity model.Application
	if err := r.collection.FindOne(ctx, bson.M{
		"_id": _id,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

func (r *ApplicationRepoerMongo) FindByName(ctx context.Context, name string) (*model.Application, error) {
	var entity model.Application
	if err := r.collection.FindOne(ctx, bson.M{
		"name": name,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

func (r *ApplicationRepoerMongo) FindByNameAndOwner(ctx context.Context, name, owner string) (*model.Application, error) {
	var entity model.Application
	if err := r.collection.FindOne(ctx, bson.M{
		"name":  name,
		"owner": owner,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

func (r *ApplicationRepoerMongo) FindByContainerID(ctx context.Context, containerID string) (*model.Application, error) {
	var entity model.Application
	if err := r.collection.FindOne(ctx, bson.M{
		"container.containerID": containerID,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

func (r *ApplicationRepoerMongo) FindByOwner(ctx context.Context, owner string) ([]*model.Application, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"owner": owner,
	}, options.Find().SetSort(bson.M{}))
	if err != nil {
		return nil, err
	}
	var entities []*model.Application
	if err := cursor.All(ctx, &entities); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return entities, nil
}

func (r *ApplicationRepoerMongo) FindByOwnerAndTypeAndIsPublicTrue(ctx context.Context, owner string, serviceType model.ServiceKind) ([]*model.Application, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"$and": []bson.M{
			{"owner": owner},
			{"type": serviceType},
			{"isPublic": true},
		},
	}, options.Find().SetSort(bson.M{}))
	if err != nil {
		return nil, err
	}
	var entities []*model.Application
	if err := cursor.All(ctx, &entities); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return entities, nil
}

func (r *ApplicationRepoerMongo) FindByOwnerAndTypeAndIsPublicFalse(ctx context.Context, owner string, serviceType model.ServiceKind) ([]*model.Application, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"$and": []bson.M{
			{"owner": owner},
			{"type": serviceType},
			{"isPublic": false},
		},
	}, options.Find().SetSort(bson.M{}))
	if err != nil {
		return nil, err
	}
	var entities []*model.Application
	if err := cursor.All(ctx, &entities); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return entities, nil
}

func (r *ApplicationRepoerMongo) FindByOwnerAndIsUpdatableTrue(ctx context.Context, owner string) ([]*model.Application, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"$and": []bson.M{
			{"owner": owner},
			{"isUpdatable": true},
		},
	}, options.Find().SetSort(bson.M{}))
	if err != nil {
		return nil, err
	}
	var entities []*model.Application
	if err := cursor.All(ctx, &entities); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return entities, nil
}

func (r *ApplicationRepoerMongo) InsertOne(ctx context.Context, application *model.Application) (interface{}, error) {
	result, err := r.collection.InsertOne(ctx, application)
	if err != nil {
		return nil, err
	}

	return result.InsertedID, nil
}

func (r *ApplicationRepoerMongo) UpdateByID(ctx context.Context, application *model.Application, _id primitive.ObjectID) (bool, error) {
	result, err := r.collection.UpdateOne(ctx, bson.M{
		"_id": _id,
	}, bson.M{
		"$set": application,
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, repo.ErrNotFound
		}
		return false, err
	}
	return result.MatchedCount > 0, err
}

func (r *ApplicationRepoerMongo) DeleteByID(ctx context.Context, _id primitive.ObjectID) (bool, error) {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"_id": _id,
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, repo.ErrNotFound
		}
		return false, err
	}
	return result.DeletedCount > 0, nil
}
