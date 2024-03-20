package k8smanager

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (k *K8sOrchestratedServiceManager) GetEventsChan(ctx context.Context) (watch.Interface, error) {
	watcher, err := k.clientset.CoreV1().Pods("").Watch(ctx,
		v1.ListOptions{
			LabelSelector: "ipaasManaged=true",
		})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}
