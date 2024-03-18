package serviceManager

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
)

var (
	ErrorCreatingResource error = fmt.Errorf("erorr")
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
	//*namespaces
	// GetNamespace(ctx context.Context, namespace string) (*model.)
	CreateNewNamespace(ctx context.Context, namespace string, labels []model.KeyValue) error
	//gracePeriod in seconds, if 0 it will be deleted immediately, if negative it will be deleted after the default grace period
	DeleteNamespace(ctx context.Context, namespace string, gracePeriod int64) error
	//to be used for creating a new namespace and then creating a new secret in it
	CreateNewRegistrySecret(ctx context.Context, namespace, registryUrl, username, password string) (string, error)

	//*deployments
	GetDeployment(ctx context.Context, namespace, deploymentName string) (*model.Deployment, error)
	CreateNewDeployment(ctx context.Context, namespace, deploymentName, app, imageRegistry string, replicas, port int32, labels []model.KeyValue, configMapName string, volume *model.Volume) (*model.Deployment, error)
	RestartDeployment(ctx context.Context, namespace, deploymentName string) error
	//gracePeriod in seconds, if 0 it will be deleted immediately, if negative it will be deleted after the default grace period
	DeleteDeployment(ctx context.Context, namespace, deploymentName string, gracePeriod int64) error
	WaitDeploymentReadyState(ctx context.Context, namespace, deploymentName string) (chan struct{}, chan error)
	//! review
	// GetRevisions(ctx context.Context, namespace, deploymentName string) ([]model.Deployment, error)
	// UpdateDeployment(ctx context.Context, namespace, deploymentName, imageRegistry string, replicas, port int32, labels []model.KeyValue, configMapName string) (*model.Deployment, error)
	// RollbackDeployment(ctx context.Context, namespace, deploymentName string, revision int64) error

	//*services
	GetService(ctx context.Context, namespace, serviceName string) (*model.Service, error)
	CreateNewService(ctx context.Context, namespace, serviceName, app string, port int32, labels []model.KeyValue) (*model.Service, error)
	UpdateService(ctx context.Context, namespace, serviceName string, port int32) error
	//gracePeriod in seconds, if 0 it will be deleted immediately, if negative it will be deleted after the default grace period
	DeleteService(ctx context.Context, namespace, serviceName string, gracePeriod int64) error

	//*ingressRoute
	GetIngressRoute(ctx context.Context, namespace, ingressRouteName string) (*model.IngressRoute, error)
	CreateNewIngressRoute(ctx context.Context, namespace, ingressRouteName, host, serviceName string, listeningPort int32, labels []model.KeyValue) (*model.IngressRoute, error)
	UpdateIngressRoute(ctx context.Context, namespace, ingressRouteName, newHost string) error
	//gracePeriod in seconds, if 0 it will be deleted immediately, if negative it will be deleted after the default grace period
	DeleteIngressRoute(ctx context.Context, namespace, ingressRouteName string, gracePeriod int64) error

	//*configMap
	GetConfigMap(ctx context.Context, namespace, configMapName string) (*model.ConfigMap, error)
	CreateNewConfigMap(ctx context.Context, namespace, configMapName string, data, labels []model.KeyValue) (*model.ConfigMap, error)
	UpdateConfigMap(ctx context.Context, namespace, configMapName string, data []model.KeyValue) error
	//gracePeriod in seconds, if 0 it will be deleted immediately, if negative it will be deleted after the default grace period
	DeleteConfigMap(ctx context.Context, namespace, configMapName string, gracePeriod int64) error

	//*pv and pvc
	CreateNewPersistentVolumeClaim(ctx context.Context, namespace, pvcName, storageClassName string, storageSize int64, labels []model.KeyValue) (*model.PersistentVolumeClaim, error)
}
