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

func NewUserRepoer(collection *mongo.Collection) repo.UserRepoer {
	return &UserRepoerMongo{
		collection: collection,
	}
}

type UserRepoerMongo struct {
	collection *mongo.Collection
}

func (r *UserRepoerMongo) InsertOne(ctx context.Context, user *model.User) (interface{}, error) {
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

func (r *UserRepoerMongo) UpdateGithubAccessTokenByID(ctx context.Context, githubAccessToken string, id primitive.ObjectID) (bool, error) {
	result, err := r.collection.UpdateOne(ctx, bson.M{
		"_id": id,
	}, bson.M{
		"$set": bson.M{
			"githubAccessToken": githubAccessToken,
		},
	})
	if err != nil {
		return false, err
	}
	return result.ModifiedCount > 0, nil
}

func (r *UserRepoerMongo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var entity model.User
	if err := r.collection.FindOne(ctx, bson.M{
		"_id": id,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *UserRepoerMongo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var entity model.User
	if err := r.collection.FindOne(ctx, bson.M{
		"email": email,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *UserRepoerMongo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var entity model.User
	if err := r.collection.FindOne(ctx, bson.M{
		"username": username,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *UserRepoerMongo) DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error) {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"_id": id,
	})
	if err != nil {
		return false, err
	}
	return result.DeletedCount > 0, nil
}
