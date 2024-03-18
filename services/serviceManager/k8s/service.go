package k8smanager

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (k K8sOrchestratedServiceManager) GetService(ctx context.Context, namespace, serviceName string) (*model.Service, error) {
	service, err := k.clientset.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting service: %v", err)
	}
	return &model.Service{
		BaseResource: model.BaseResource{
			Name:      service.Name,
			Namespace: service.Namespace,
			Labels:    convertK8sDataToModelData(service.Labels),
		},
		Port:       service.Spec.Ports[0].Port,
		TargetPort: service.Spec.Ports[0].TargetPort.IntVal,
	}, nil

}

func (k K8sOrchestratedServiceManager) CreateNewService(ctx context.Context, namespace, serviceName, app string, port int32, labels []model.KeyValue) (*model.Service, error) {
	k8sLabels := convertModelKeyValuesToLables(labels)
	_, err := k.clientset.CoreV1().Services(namespace).
		Create(ctx,
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:   serviceName,
					Labels: k8sLabels,
				},
				Spec: v1.ServiceSpec{
					Selector: map[string]string{
						model.AppLabel: app,
					},
					Ports: []v1.ServicePort{
						{
							Port: port,
							TargetPort: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: port,
							},
						},
					},
				},
			}, metav1.CreateOptions{})

	if err != nil {
		return nil, fmt.Errorf("error creating deployment: %v", err)
	}
	return &model.Service{
		BaseResource: model.BaseResource{
			Name:      serviceName,
			Namespace: namespace,
			Labels:    labels,
		},
		Port:       port,
		TargetPort: port,
	}, nil
}

func (k K8sOrchestratedServiceManager) UpdateService(ctx context.Context, namespace, serviceName string, port int32) error {
	service, err := k.clientset.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting service: %v", err)
	}
	service.Spec.Ports[0].Port = port
	service.Spec.Ports[0].TargetPort = intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: port,
	}
	_, err = k.clientset.CoreV1().Services(namespace).Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating service: %v", err)
	}
	return nil
}

func (k K8sOrchestratedServiceManager) DeleteService(ctx context.Context, namespace, serviceName string, gracePeriod int64) error {
	err := k.clientset.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	})
	if err != nil {
		return fmt.Errorf("error deleting service: %v", err)
	}
	return nil
}
