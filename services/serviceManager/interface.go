package serviceManager

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

// TODO: volume support is not implemented in
// this version
type ServiceManager interface {
	CreateNewService(ctx context.Context, name, image string, envs, labels []model.KeyValue) (*model.Container, error)
	StartServiceByID(ctx context.Context, id string) error
	//force remove a container
	RemoveServiceByID(ctx context.Context, id string, force bool) error
	StopServiceByID(ctx context.Context, id string) error

	CreateNewNetwork(ctx context.Context, name string) (string, error)
	RemoveNetwork(ctx context.Context, id string) error
	ConnectServiceToNetwork(ctx context.Context, id, networkID, dnsAlias string) error

	RemoveImageByID(ctx context.Context, id string) error
}
