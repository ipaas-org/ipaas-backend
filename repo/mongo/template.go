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
	var template model.Template
	if err := r.collection.FindOne(ctx, bson.M{
		"code": code,
	}, options.FindOne().SetSort(bson.M{})).Decode(&template); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepoerMongo) FindAll(ctx context.Context) ([]*model.Template, error) {
	var templates []*model.Template
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{}))
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *TemplateRepoerMongo) FindAllAvailable(ctx context.Context) ([]*model.Template, error) {
	var templates []*model.Template
	cursor, err := r.collection.Find(ctx, bson.M{
		"available": true,
	}, options.Find().SetSort(bson.M{}))
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}
