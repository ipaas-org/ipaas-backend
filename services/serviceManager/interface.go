package serviceManager

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

// TODO: volume support is not implemented in
// this version
type ServiceManager interface {
	CreateNewContainer(ctx context.Context, image string, envs, labels []model.KeyValue) (*model.Container, error)
	StartContainer(ctx context.Context, container *model.Container) error
	//force remove a container
	RemoveContainer(ctx context.Context, container *model.Container) error

	CreateNewNetwork(ctx context.Context, name string) (string, error)
	ConnectContainerToNetwork(ctx context.Context, container *model.Container, networkID, dnsAlias string) error
	RemoveNetwork(ctx context.Context, id string) error
}
