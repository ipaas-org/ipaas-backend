package serviceManager

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

// TODO: volume support is not implemented in
// this version
type ServiceManager interface {
	//create a container and return id and container name
	CreateNewContainer(ctx context.Context, name, image string, envs, labels []model.KeyValue) (string, string, error)

	ConnectContainerToNetwork(ctx context.Context, containerID, networkID, dnsAlias string) error
	//force remove a container
	RemoveContainer(ctx context.Context, id string) error

	StartContainer(ctx context.Context, id string) error
}
