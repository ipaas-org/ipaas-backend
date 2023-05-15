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

func (r *ApplicationRepoerMongo) FindByName(arg0 context.Context, arg1 string) (*model.Application, error) {
	var entity model.Application
	if err := r.collection.FindOne(arg0, bson.M{
		"name": arg1,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *ApplicationRepoerMongo) FindByNameAndOwnerUsername(ctx context.Context, name, ownerUsername string) (*model.Application, error) {
	var entity model.Application
	if err := r.collection.FindOne(ctx, bson.M{
		"name":          name,
		"ownerUsername": ownerUsername,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *ApplicationRepoerMongo) FindByContainerID(arg0 context.Context, arg1 string) (*model.Application, error) {
	var entity model.Application
	if err := r.collection.FindOne(arg0, bson.M{
		"containerID": arg1,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *ApplicationRepoerMongo) FindByOwnerUsername(arg0 context.Context, arg1 string) ([]*model.Application, error) {
	cursor, err := r.collection.Find(arg0, bson.M{
		"ownerUsername": arg1,
	}, options.Find().SetSort(bson.M{}))
	if err != nil {
		return nil, err
	}
	var entities []*model.Application
	if err := cursor.All(arg0, &entities); err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *ApplicationRepoerMongo) FindByOwnerUsernameAndTypeAndIsPublicTrue(arg0 context.Context, arg1 string, arg2 string) ([]*model.Application, error) {
	cursor, err := r.collection.Find(arg0, bson.M{
		"$and": []bson.M{
			{"ownerUsername": arg1},
			{"type": arg2},
			{"isPublic": true},
		},
	}, options.Find().SetSort(bson.M{}))
	if err != nil {
		return nil, err
	}
	var entities []*model.Application
	if err := cursor.All(arg0, &entities); err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *ApplicationRepoerMongo) FindByOwnerUsernameAndTypeAndIsPublicFalse(arg0 context.Context, arg1 string, arg2 string) ([]*model.Application, error) {
	cursor, err := r.collection.Find(arg0, bson.M{
		"$and": []bson.M{
			{"ownerUsername": arg1},
			{"type": arg2},
			{"isPublic": false},
		},
	}, options.Find().SetSort(bson.M{}))
	if err != nil {
		return nil, err
	}
	var entities []*model.Application
	if err := cursor.All(arg0, &entities); err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *ApplicationRepoerMongo) FindByOwnerUsernameAndIsUpdatableTrue(arg0 context.Context, arg1 string) ([]*model.Application, error) {
	cursor, err := r.collection.Find(arg0, bson.M{
		"$and": []bson.M{
			{"ownerUsername": arg1},
			{"isUpdatable": true},
		},
	}, options.Find().SetSort(bson.M{}))
	if err != nil {
		return nil, err
	}
	var entities []*model.Application
	if err := cursor.All(arg0, &entities); err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *ApplicationRepoerMongo) Insert(arg0 context.Context, arg1 *model.Application) (interface{}, error) {
	result, err := r.collection.InsertOne(arg0, arg1)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

func (r *ApplicationRepoerMongo) UpdateByID(arg0 context.Context, arg1 *model.Application, arg2 primitive.ObjectID) (bool, error) {
	result, err := r.collection.UpdateOne(arg0, bson.M{
		"_id": arg2,
	}, bson.M{
		"$set": arg1,
	})
	if err != nil {
		return false, err
	}
	return result.MatchedCount > 0, err
}

func (r *ApplicationRepoerMongo) DeleteByID(arg0 context.Context, arg1 primitive.ObjectID) (bool, error) {
	result, err := r.collection.DeleteOne(arg0, bson.M{
		"_id": arg1,
	})
	if err != nil {
		return false, err
	}
	return result.DeletedCount > 0, nil
}
