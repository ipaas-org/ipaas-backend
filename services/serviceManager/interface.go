package serviceManager

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

type ServiceManager interface {
	//create a container and return id and container name
	CreateNewContainer(ctx context.Context, image string, envs []model.Env) (string, string, error)
	//force remove a container
	RemoveContainer(ctx context.Context, id string) error

	StartContainer(ctx context.Context, id string) error
}
