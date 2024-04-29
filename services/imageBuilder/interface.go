package imageBuilder

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

type ImageBuilder interface {
	BuildImage(ctx context.Context, buildInfo model.Request) error
	ValidateImageResponse(response model.BuildResponse) (string, error)
}
