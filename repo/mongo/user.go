package mongo

import (
	"context"
	"time"

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
	t := time.Now()
	user.CreatedAt = t
	user.UpdatedAt = t
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	return result.InsertedID, nil
}

func (r *UserRepoerMongo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var user model.User
	if err := r.collection.FindOne(ctx, bson.M{
		"_id": id,
	}, options.FindOne().SetSort(bson.M{})).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepoerMongo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.collection.FindOne(ctx, bson.M{
		"userInfo.email": email,
	}, options.FindOne().SetSort(bson.M{})).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepoerMongo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.collection.FindOne(ctx, bson.M{
		"userInfo.username": username,
	}, options.FindOne().SetSort(bson.M{})).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepoerMongo) FindByCode(ctx context.Context, userCode string) (*model.User, error) {
	var user model.User
	if err := r.collection.FindOne(ctx, bson.M{
		"code": userCode,
	}, options.FindOne().SetSort(bson.M{})).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepoerMongo) UpdateGithubAccessTokenByID(ctx context.Context, githubAccessToken string, id primitive.ObjectID) (bool, error) {
	result, err := r.collection.UpdateOne(ctx, bson.M{
		"_id": id,
	}, bson.M{
		"$set": bson.M{
			"userInfo. githubAccessToken": githubAccessToken,
			"updatedAt":                   time.Now(),
		},
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, repo.ErrNotFound
		}
		return false, err
	}
	return result.ModifiedCount > 0, nil
}

func (r *UserRepoerMongo) DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error) {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"_id": id,
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, repo.ErrNotFound
		}
		return false, err
	}
	return result.DeletedCount > 0, nil
}
