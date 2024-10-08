package k8s

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ipaas-org/ipaas-backend/model"
	k8smanager "github.com/ipaas-org/ipaas-backend/services/serviceManager/k8s"
)

var defaultLabels = []model.KeyValue{
	{
		Key:   model.IpaasManagedLabel,
		Value: "true",
	},
	{
		Key:   model.EnvironmentLabel,
		Value: "test",
	},
}

func getTestK8sManager() *k8smanager.K8sOrchestratedServiceManager {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	manager, err := k8smanager.NewK8sOrchestratedServiceManager(home+"/.kube/config", "100m", "100Mi")
	if err != nil {
		panic(err)
	}
	return manager
}

func TestCreateNewNamespace(t *testing.T) {
	manager := getTestK8sManager()
	namespace := "test-namespace"
	ctx := context.Background()
	_, err := manager.CreateNewNamespace(ctx, namespace, defaultLabels)
	if err != nil {
		t.Errorf("error creating namespace: %v\n", err)
	}

	t.Cleanup(func() {
		err := manager.DeleteNamespace(ctx, namespace, 0)
		if err != nil {
			t.Errorf("error deleting namespace: %v\n", err)
		}
	})
}

func TestCreateConfigMap(t *testing.T) {
	manager := getTestK8sManager()
	namespace := "test-namespace"
	configMapName := "test-configmap"
	ctx := context.Background()
	_, err := manager.CreateNewNamespace(ctx, namespace, defaultLabels)
	if err != nil {
		t.Errorf("error creating namespace: %v\n", err)
	}

	data := []model.KeyValue{
		{
			Key:   "test-key",
			Value: "test-value",
		},
		{
			Key: "test-file",
			Value: `this is 
some very very big content
which is on multiple lines`,
		},
	}
	labels := append(defaultLabels, model.KeyValue{
		Key:   model.ResourceNameLabel,
		Value: configMapName,
	})
	configMap, err := manager.CreateNewConfigMap(ctx, namespace, configMapName, data, labels)
	if err != nil {
		t.Errorf("error creating configmap: %v\n", err)
	}
	t.Log("configmap created: ", configMap)
	<-time.After(10 * time.Second)
	t.Cleanup(func() {
		err := manager.DeleteNamespace(ctx, namespace, 0)
		if err != nil {
			t.Errorf("error deleting namespace: %v\n", err)
		}
	})
}

func TestCreateNewDeployment(t *testing.T) {
	manager := getTestK8sManager()
	namespace := "test-namespace"
	// realImage := "registry.cargoway.cloud/us-60ff775c-4915-4523-bf91-ccb63874d95d/65e8ab54c3086ba65b91cb16:22e7f5b80c0e537a72d09dae4de7ab245c5ac8f9"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := manager.CreateNewNamespace(ctx, namespace, defaultLabels)
	if err != nil {
		t.Fatalf("error creating namespace: %v\n", err)
	}

	done, errChan := manager.WaitDeploymentReadyState(ctx, namespace, "test-deployment")

	data := []model.KeyValue{
		{
			Key:   "test-key",
			Value: "test-value",
		},
		{
			Key: "test-file",
			Value: `this is 
some very very big content
which is on multiple lines`,
		},
	}
	configMap, err := manager.CreateNewConfigMap(ctx, namespace, "test-configmap", data, defaultLabels)
	if err != nil {
		t.Fatalf("error creating configmap: %v\n", err)
	}
	t.Log("configmap created: ", configMap)

	labels := append(defaultLabels, model.KeyValue{
		Key:   model.ResourceNameLabel,
		Value: "test-deployment",
	})
	dep, err := manager.CreateNewDeployment(ctx, namespace, "test-deployment", "nginx-test", "ubuntu/nginx", 1, 80, labels, configMap.Name, nil)
	if err != nil {
		t.Fatalf("error creating deployment: %v\n", err)
	}
	t.Logf("deployment created: %+v", dep)

loop:
	for {
		select {
		case err := <-errChan:
			t.Log("found error while waiting for deployment", err)
			break loop
		case _, ok := <-done:
			if !ok {
				t.Log("done chan is closed, either context was cancelled or something worst O-O")
			} else {
				t.Log("found the done bitch")
			}
			break loop
		}
	}

	t.Cleanup(func() {
		t.Log("removing namespace", namespace)
		err := manager.DeleteNamespace(ctx, namespace, 0)
		if err != nil {
			t.Errorf("error deleting namespace: %v\n", err)
		}
	})
}

func TestRestartDeployment(t *testing.T) {
	manager := getTestK8sManager()
	namespace := "test-namespace"
	ctx := context.Background()
	_, err := manager.CreateNewNamespace(ctx, namespace, defaultLabels)
	if err != nil {
		t.Errorf("error creating namespace: %v\n", err)
	}

	data := []model.KeyValue{
		{
			Key:   "test-key",
			Value: "test-value",
		},
	}
	configMap, err := manager.CreateNewConfigMap(ctx, namespace, "test-configmap", data, defaultLabels)
	if err != nil {
		t.Errorf("error creating configmap: %v\n", err)
	}
	t.Log("configmap created: ", configMap)

	dep, err := manager.CreateNewDeployment(ctx, namespace, "test-deployment", "nginx-test", "ubuntu/nginx", 1, 80, defaultLabels, configMap.Name, nil)
	if err != nil {
		t.Errorf("error creating deployment: %v\n", err)
	}
	t.Logf("deployment created: %+v", dep)

	time.Sleep(10 * time.Second)

	err = manager.RestartDeployment(ctx, namespace, dep.Name)
	if err != nil {
		t.Errorf("error restarting deployment: %v\n", err)
	}

	// t.Cleanup(func() {
}

func TestDeleteNamespace(t *testing.T) {
	manager := getTestK8sManager()
	namespace := "test-namespace"
	ctx := context.Background()
	// err := manager.DeleteNamespace(ctx, namespace, 0)
	// if err != nil {
	// 	t.Errorf("error creating namespace: %v\n", err)
	// }
	namespaceLabels := append(defaultLabels, model.KeyValue{
		Key:   model.ResourceNameLabel,
		Value: namespace,
	})

	done, errChan := manager.WaitForNamespaceRemoval(ctx, namespace)

	if err := manager.DeleteNamespace(ctx, namespace, 0); err != nil {
		t.Errorf("error deleting namespace: %v\n", err)
	}

loop:
	for {
		select {
		case err := <-errChan:
			t.Log("found error while waiting for namespace removal:", err)
			break loop
		case _, ok := <-done:
			if !ok {
				t.Log("done chan is closed, either context was cancelled or something worst O-O")
			} else {
				t.Log("found stuff")
				_, err := manager.CreateNewNamespace(ctx, namespace, namespaceLabels)
				if err != nil {
					t.Error("incorrect behaviour of delete namespace:", err)
				}
			}
			break loop
		}
	}
}

func TestFullApplicationStartup(t *testing.T) {
	manager := getTestK8sManager()
	namespace := "test-namespace"
	image := "registry.cargoway.cloud/library/heavy:latest"
	ctx := context.Background()

	namespaceLabels := append(defaultLabels, model.KeyValue{
		Key:   model.ResourceNameLabel,
		Value: namespace,
	})
	_, err := manager.CreateNewNamespace(ctx, namespace, namespaceLabels)
	if err != nil {
		t.Errorf("error creating namespace: %v\n", err)
	}

	data := []model.KeyValue{
		{
			Key:   "test-key",
			Value: "test-value",
		},
	}

	configMap, err := manager.CreateNewConfigMap(ctx, namespace, "test-configmap", data, defaultLabels)
	if err != nil {
		t.Errorf("error creating configmap: %v\n", err)
	}
	t.Log("configmap created: ", configMap)

	dep, err := manager.CreateNewDeployment(ctx, namespace, "test-deployment", "nginx-test", image, 1, 8080, defaultLabels, configMap.Name, nil)
	if err != nil {
		t.Errorf("error creating deployment: %v\n", err)
	}
	t.Logf("deployment created: %+v", dep)

	service, err := manager.CreateNewService(ctx, namespace, "test-service", "nginx-test", 8080, defaultLabels)
	if err != nil {
		t.Errorf("error creating service: %v\n", err)
	}
	t.Logf("service created: %+v", service)

	ingress, err := manager.CreateNewIngressRoute(ctx, namespace, "test-ingress", "testing.cargoway.cloud", "test-service", 8080, defaultLabels)
	if err != nil {
		t.Errorf("error creating ingress: %v\n", err)
	}
	t.Logf("ingress created: %+v", ingress)

	// t.Cleanup(func() {
	// 	err := manager.DeleteNamespace(ctx, namespace, 0)
	// 	if err != nil {
	// 		t.Errorf("error deleting namespace: %v\n", err)
	// 	}
	// })
}

func TestNewServiceCreation(t *testing.T) {
	manager := getTestK8sManager()
	namespace := "test-namespace"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := manager.CreateNewNamespace(ctx, namespace, defaultLabels)
	if err != nil {
		t.Fatalf("error creating namespace: %v\n", err)
	}

	data := []model.KeyValue{
		{
			Key:   "test-key",
			Value: "test-value",
		},
		{
			Key: "test-file",
			Value: `this is 
some very very big content
which is on multiple lines`,
		},
	}
	configMap, err := manager.CreateNewConfigMap(ctx, namespace, "test-configmap", data, defaultLabels)
	if err != nil {
		t.Fatalf("error creating configmap: %v\n", err)
	}
	t.Log("configmap created: ", configMap)

	labels := append(defaultLabels, model.KeyValue{
		Key:   model.ResourceNameLabel,
		Value: "test-service",
	})
	dep, err := manager.CreateNewDeployment(ctx, namespace, "test-deployment", "nginx-test", "ubuntu/nginx", 1, 80, defaultLabels, configMap.Name, nil)
	if err != nil {
		t.Fatalf("error creating deployment: %v\n", err)
	}
	t.Logf("deployment created: %+v", dep)

	svc, err := manager.CreateNewService(ctx, namespace, "test-service", "nginx-test", 80, labels)
	if err != nil {
		t.Fatalf("error creating service: %v\n", err)
	}
	t.Logf("service created: %+v", svc)

	t.Cleanup(func() {
		ctx := context.Background()
		t.Log("removing namespace", namespace)
		err := manager.DeleteNamespace(ctx, namespace, 0)
		if err != nil {
			t.Errorf("error deleting namespace: %v\n", err)
		}
	})
}
