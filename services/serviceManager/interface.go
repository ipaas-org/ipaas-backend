package serviceManager

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

// TODO: volume support is not implemented in
// this version
type ServiceManager interface {
	CreateNewContainer(ctx context.Context, name, image string, envs, labels []model.KeyValue) (*model.Container, error)
	StartContainerByID(ctx context.Context, id string) error
	//force remove a container
	RemoveContainerByID(ctx context.Context, id string, force bool) error
	StopContainerByID(ctx context.Context, id string) error

	CreateNewNetwork(ctx context.Context, name string) (string, error)
	RemoveNetwork(ctx context.Context, id string) error
	ConnectContainerToNetwork(ctx context.Context, id, networkID, dnsAlias string) error

	RemoveImageByID(ctx context.Context, id string) error
}
