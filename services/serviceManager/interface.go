package serviceManager

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

// TODO: volume support is not implemented in
// this version
type ServiceManager interface {
	CreateNewService(ctx context.Context, name, image string, envs, labels []model.KeyValue) (*model.Service, error)
	StartServiceByID(ctx context.Context, id string) error
	//force remove a container
	RemoveServiceByID(ctx context.Context, id string, force bool) error
	StopServiceByID(ctx context.Context, id string) error

	CreateNewNetwork(ctx context.Context, name string) (string, error)
	RemoveNetwork(ctx context.Context, id string) error
	ConnectServiceToNetwork(ctx context.Context, id, networkID, dnsAlias string) error

	RemoveImageByID(ctx context.Context, id string) error
}

type OrchestratedServiceManager interface {
	CreateNewNamespace(ctx context.Context, namespace, owner, environment string) error
	CreateDeployment(ctx context.Context, namespace, deploymentName, app, owner, environment, visibility, imageRegistry string, replicas, port int32, envs []model.KeyValue) error
}
