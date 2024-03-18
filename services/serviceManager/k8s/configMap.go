package k8smanager

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k K8sOrchestratedServiceManager) CreateNewConfigMap(ctx context.Context, namespace, configMapName string, data, labels []model.KeyValue) (*model.ConfigMap, error) {
	k8sLabels := convertModelKeyValuesToLables(labels)
	k.clientset.CoreV1().ConfigMaps(namespace).Create(ctx,
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapName,
				Labels:    k8sLabels,
				Namespace: namespace,
			},
			Data: convertModelDataToK8sData(data),
		}, metav1.CreateOptions{})

	return &model.ConfigMap{
		BaseResource: model.BaseResource{
			Name:      configMapName,
			Namespace: namespace,
			Labels:    labels,
		},
	}, nil
}

func (k K8sOrchestratedServiceManager) UpdateConfigMap(ctx context.Context, namespace, configMapName string, data []model.KeyValue) error {
	// k8sLabels := convertModelKeyValuesToLables(labels)
	k.clientset.CoreV1().ConfigMaps(namespace).Update(ctx,
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: configMapName,
				// Labels:    k8sLabels,
				Namespace: namespace,
			},
			Data: convertModelDataToK8sData(data),
		}, metav1.UpdateOptions{})

	return nil
}

func (k K8sOrchestratedServiceManager) GetConfigMap(ctx context.Context, namespace, configMapName string) (*model.ConfigMap, error) {
	configMap, err := k.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting config map: %v", err)
	}
	return &model.ConfigMap{
		BaseResource: model.BaseResource{
			Name:      configMap.Name,
			Namespace: configMap.Namespace,
			Labels:    convertK8sDataToModelData(configMap.Labels),
		},
		Data: convertK8sDataToModelData(configMap.Data),
	}, nil
}

func (k K8sOrchestratedServiceManager) DeleteConfigMap(ctx context.Context, namespace, configMapName string, gracePeriod int64) error {
	grace := &gracePeriod
	if gracePeriod < 0 {
		grace = nil
	}
	err := k.clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, configMapName, metav1.DeleteOptions{
		GracePeriodSeconds: grace,
	})
	if err != nil {
		return fmt.Errorf("error deleting config map: %v", err)
	}
	return nil
}
