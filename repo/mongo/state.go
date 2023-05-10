package mongo

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewStateRepoer(collection *mongo.Collection) repo.StateRepoer {
	return &StateRepoerMongo{
		collection: collection,
	}
}

type StateRepoerMongo struct {
	collection *mongo.Collection
}

func (r *StateRepoerMongo) FindByState(ctx context.Context, state string) (*model.State, error) {
	var entity model.State
	if err := r.collection.FindOne(ctx, bson.M{
		"state": state,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *StateRepoerMongo) Insert(ctx context.Context, state *model.State) (interface{}, error) {
	result, err := r.collection.InsertOne(ctx, state)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

func (r *StateRepoerMongo) DeleteByState(ctx context.Context, state string) (bool, error) {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"state": state,
	})
	if err != nil {
		return false, err
	}
	return result.DeletedCount > 0, nil
}
