package k8smanager

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ipaas-org/ipaas-backend/model"
)

// 1 * 1024 * 1024 * 1024 = 1Gi
func (k K8sOrchestratedServiceManager) CreateNewPersistentVolumeClaim(ctx context.Context, namespace, pvcName, storageClassName string, storageSize int64, labels []model.KeyValue) (*model.PersistentVolumeClaim, error) {
	k8sLabels := convertModelKeyValuesToLables(labels)
	storageQuantity := resource.NewQuantity(storageSize, resource.BinarySI)
	fmt.Printf("resource %v\n", *storageQuantity)
	pvc, err := k.clientset.CoreV1().PersistentVolumeClaims(namespace).
		Create(ctx,
			&corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:   pvcName,
					Labels: k8sLabels,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: &storageClassName,
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: *storageQuantity,
						},
						Limits: corev1.ResourceList{
							corev1.ResourceStorage: *storageQuantity,
						},
					},
				},
			}, metav1.CreateOptions{})

	if err != nil {
		return nil, err
	}
	return &model.PersistentVolumeClaim{
		BaseResource: model.BaseResource{
			Name:      pvc.Name,
			Namespace: pvc.Namespace,
			Labels:    convertK8sDataToModelData(pvc.Labels),
		},
		StorageClassName: storageClassName,
		AccessModes:      string(corev1.ReadWriteOncePod),
		StorageSize:      pvc.Spec.Resources.Limits.Storage().Value(),
	}, nil
}
