package mongo

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewTemplateRepoer(collection *mongo.Collection) repo.TemplateRepoer {
	return &TemplateRepoerMongo{
		collection: collection,
	}
}

type TemplateRepoerMongo struct {
	collection *mongo.Collection
}

func (r *TemplateRepoerMongo) FindByCode(ctx context.Context, code string) (*model.Template, error) {
	var entity model.Template
	if err := r.collection.FindOne(ctx, bson.M{
		"code": code,
	}, options.FindOne().SetSort(bson.M{})).Decode(&entity); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}
