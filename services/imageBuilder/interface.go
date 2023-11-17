package imageBuilder

import "github.com/ipaas-org/ipaas-backend/model"

type ImageBuilder interface {
	BuildImage(buildInfo model.BuildRequest) error
	ValidateImageResponse(response model.BuildResponse) (string, error)
}
