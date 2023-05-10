package mongo

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewTokenRepoer(collection *mongo.Collection) repo.TokenRepoer {
	return &TokenRepoerMongo{
		collection: collection,
	}
}

type TokenRepoerMongo struct {
	collection *mongo.Collection
}

func (r *TokenRepoerMongo) FindByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var entity model.RefreshToken
	if err := r.collection.FindOne(ctx, bson.M{
		"token": token,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *TokenRepoerMongo) Insert(ctx context.Context, token *model.RefreshToken) (interface{}, error) {
	result, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

func (r *TokenRepoerMongo) DeleteByToken(ctx context.Context, token string) (bool, error) {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"token": token,
	})
	if err != nil {
		return false, err
	}
	return result.DeletedCount > 0, nil
}
