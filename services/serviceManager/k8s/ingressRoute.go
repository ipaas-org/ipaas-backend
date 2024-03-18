package k8smanager

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (k K8sOrchestratedServiceManager) GetIngressRoute(ctx context.Context, namespace, ingressRouteName string) (*model.IngressRoute, error) {
	ingressRoute, err := k.traefikClient.
		TraefikV1alpha1().
		IngressRoutes(namespace).
		Get(ctx, ingressRouteName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting ingress route: %v", err)
	}
	return &model.IngressRoute{
		BaseResource: model.BaseResource{
			Name:      ingressRoute.Name,
			Namespace: ingressRoute.Namespace,
			Labels:    convertK8sDataToModelData(ingressRoute.Labels),
		},
		Entrypoints: ingressRoute.Spec.EntryPoints,
		Domain:      ingressRoute.Spec.Routes[0].Match,
	}, nil
}

func (k K8sOrchestratedServiceManager) CreateNewIngressRoute(ctx context.Context, namespace, ingressRouteName, host, serviceName string, listeningPort int32, labels []model.KeyValue) (*model.IngressRoute, error) {
	k8sLables := convertModelKeyValuesToLables(labels)
	entrypoints := []string{"web", "websecure"}
	_, err := k.traefikClient.
		TraefikV1alpha1().
		IngressRoutes(namespace).
		Create(ctx,
			&traefikv1alpha1.IngressRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ingressRouteName,
					Labels:    k8sLables,
					Namespace: namespace,
				},
				Spec: traefikv1alpha1.IngressRouteSpec{
					EntryPoints: entrypoints,
					Routes: []traefikv1alpha1.Route{
						{
							Match: fmt.Sprintf("Host(`%s`)", host),
							Kind:  "Rule",
							Services: []traefikv1alpha1.Service{
								{
									LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
										Name: serviceName,
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: listeningPort,
										},
									},
								},
							},
						},
					},
				},
			},
			metav1.CreateOptions{})

	if err != nil {
		return nil, fmt.Errorf("error creating ingress route: %v", err)
	}
	return &model.IngressRoute{
		BaseResource: model.BaseResource{
			Name:      ingressRouteName,
			Namespace: namespace,
			Labels:    labels,
		},
		Entrypoints: entrypoints,
		Domain:      host,
	}, nil
}

func (k K8sOrchestratedServiceManager) UpdateIngressRoute(ctx context.Context, namespace, ingressRouteName, newHost string) error {
	ingressRoute, err := k.traefikClient.
		TraefikV1alpha1().
		IngressRoutes(namespace).
		Get(ctx, ingressRouteName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting ingress route: %v", err)
	}
	ingressRoute.Spec.Routes[0].Match = fmt.Sprintf("Host(`%s`)", newHost)
	_, err = k.traefikClient.
		TraefikV1alpha1().
		IngressRoutes(namespace).
		Update(ctx, ingressRoute, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating ingress route: %v", err)
	}
	return nil
}

func (k K8sOrchestratedServiceManager) DeleteIngressRoute(ctx context.Context, namespace, ingressRouteName string, gracePeriod int64) error {
	grace := &gracePeriod
	if gracePeriod < 0 {
		grace = nil
	}
	err := k.traefikClient.
		TraefikV1alpha1().
		IngressRoutes(namespace).
		Delete(ctx, ingressRouteName, metav1.DeleteOptions{
			GracePeriodSeconds: grace,
		})
	if err != nil {
		return fmt.Errorf("error deleting ingress route: %v", err)
	}
	return nil
}
