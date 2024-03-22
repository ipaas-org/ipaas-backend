package k8smanager

import (
	"fmt"

	"github.com/ipaas-org/ipaas-backend/services/serviceManager"
	traefikv "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var _ serviceManager.OrchestratedServiceManager = new(K8sOrchestratedServiceManager)

type K8sOrchestratedServiceManager struct {
	clientset      kubernetes.Interface
	kubeConfig     *rest.Config
	traefikClient  *traefikv.Clientset
	cpuResource    resource.Quantity
	memoryResource resource.Quantity
}

func NewK8sOrchestratedServiceManager(kubeConfigPath, cpuResource, memoryResource string) (*K8sOrchestratedServiceManager, error) {
	var config *rest.Config
	var err error
	if kubeConfigPath == "inside" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("fail to build the k8s config. Error - %s", err)
		}

	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("error getting kubernetes config: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes config: %v", err)
	}

	traefikClient, err := traefikv.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating traefik client: %v", err)
	}

	cpuQuantity, err := resource.ParseQuantity(cpuResource)
	if err != nil {
		return nil, fmt.Errorf("error parsing cpu resource: %v", err)
	}

	memoryQuantity, err := resource.ParseQuantity(memoryResource)
	if err != nil {
		return nil, fmt.Errorf("error parsing memory resource: %v", err)
	}
	return &K8sOrchestratedServiceManager{
		clientset:      clientset,
		kubeConfig:     config,
		traefikClient:  traefikClient,
		cpuResource:    cpuQuantity,
		memoryResource: memoryQuantity,
	}, nil
}
