package k8smanager

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *K8sOrchestratedServiceManager) DeletePod(ctx context.Context, namespace, podName string) error {
	return k.clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
}
