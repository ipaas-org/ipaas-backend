package k8smanager

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func convertK8sIngressRouteToModelIngressRoute(ingressRoute *traefikv1alpha1.IngressRoute) *model.IngressRoute {
	return &model.IngressRoute{
		BaseResource: model.BaseResource{
			Name:      ingressRoute.Name,
			Namespace: ingressRoute.Namespace,
			Labels:    convertK8sDataToModelData(ingressRoute.Labels),
		},
		Entrypoints: ingressRoute.Spec.EntryPoints,
		Match:       ingressRoute.Spec.Routes[0].Match,
	}
}

func (k K8sOrchestratedServiceManager) GetIngressRoute(ctx context.Context, namespace, ingressRouteName string) (*model.IngressRoute, error) {
	ingressRoute, err := k.traefikClient.
		TraefikV1alpha1().
		IngressRoutes(namespace).
		Get(ctx, ingressRouteName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting ingress route: %v", err)
	}
	return convertK8sIngressRouteToModelIngressRoute(ingressRoute), nil
}

func (k K8sOrchestratedServiceManager) CreateNewIngressRoute(ctx context.Context, namespace, ingressRouteName, match, serviceName string, listeningPort int32, labels []model.KeyValue) (*model.IngressRoute, error) {
	k8sLables := convertModelDataToK8sData(labels)
	entrypoints := []string{"web", "websecure"}
	ingressRoute, err := k.traefikClient.
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
							Match: match,
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
	return convertK8sIngressRouteToModelIngressRoute(ingressRoute), nil
}

func (k K8sOrchestratedServiceManager) UpdateIngressRoute(ctx context.Context, namespace, ingressRouteName, newMatch string, newPort int32) (*model.IngressRoute, error) {
	ingressRoute, err := k.traefikClient.
		TraefikV1alpha1().
		IngressRoutes(namespace).
		Get(ctx, ingressRouteName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting ingress route: %v", err)
	}
	ingressRoute.Spec.Routes[0].Match = newMatch
	ingressRoute.Spec.Routes[0].Services[0].LoadBalancerSpec.Port = intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: newPort,
	}
	_, err = k.traefikClient.
		TraefikV1alpha1().
		IngressRoutes(namespace).
		Update(ctx, ingressRoute, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error updating ingress route: %v", err)
	}
	return convertK8sIngressRouteToModelIngressRoute(ingressRoute), nil
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
