package k8smanager

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func convertKubeNamespaceToModelBaseResource(ns *v1.Namespace) *model.BaseResource {
	return &model.BaseResource{
		Name:      ns.Name,
		Namespace: ns.Namespace,
		Labels:    convertK8sDataToModelData(ns.Labels),
	}
}

func (k K8sOrchestratedServiceManager) GetNamespace(ctx context.Context, namespace string) (*model.BaseResource, error) {
	ns, err := k.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting namespace: %v", err)
	}
	return convertKubeNamespaceToModelBaseResource(ns), nil
}

func (k K8sOrchestratedServiceManager) CreateNewNamespace(ctx context.Context, namespace string, labels []model.KeyValue) (*model.BaseResource, error) {
	k8sLabels := convertModelDataToK8sData(labels)
	ns, err := k.clientset.CoreV1().Namespaces().Create(ctx,
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   namespace,
				Labels: k8sLabels,
			},
		}, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating namespace: %v", err)
	}
	return convertKubeNamespaceToModelBaseResource(ns), err
}

func (k K8sOrchestratedServiceManager) DeleteNamespace(ctx context.Context, namespace string, gracePeriod int64) error {
	err := k.clientset.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	})
	if err != nil {
		return fmt.Errorf("error deleting namespace: %v", err)
	}
	return nil
}

func (k K8sOrchestratedServiceManager) CreateNewRegistrySecret(ctx context.Context, namespace, registryUrl, username, password string) (string, error) {
	_, err := k.clientset.CoreV1().Secrets(namespace).Create(ctx,
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "registrypullsecret",
			},
			Type: v1.SecretTypeDockerConfigJson,
			StringData: map[string]string{
				".dockerconfigjson": fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"%s"}}}`,
					registryUrl, username, password)},
		}, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("error creating registry secret: %v", err)
	}
	return "registrypullsecret", nil
}

func (k K8sOrchestratedServiceManager) WaitForNamespaceRemoval(ctx context.Context, namespace string) (chan struct{}, chan error) {
	done := make(chan struct{})
	errChan := make(chan error)
	timeout := int64(5 * 60)
	deploymentWatch, err := k.clientset.CoreV1().Namespaces().Watch(ctx, metav1.ListOptions{
		LabelSelector:  model.ResourceNameLabel + "=" + namespace,
		TimeoutSeconds: &timeout,
	})
	if err != nil {
		errChan <- fmt.Errorf("error creating watch for deployment: %v", err)
		return nil, errChan
	}
	go func() {
		for {
			select {
			case e, ok := <-deploymentWatch.ResultChan():
				if !ok {
					deploymentWatch.Stop()
					errChan <- fmt.Errorf("timeout reached")
					return
				}
				if e.Type == watch.Deleted {
					done <- struct{}{}
				}

			case <-ctx.Done():
				close(done)
				return
			}
		}
	}()
	return done, errChan
}
