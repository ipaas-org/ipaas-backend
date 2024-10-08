package k8smanager

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func convertK8sMiddlewareToModelMiddleware(middleware *v1alpha1.Middleware) *model.Middleware {
	if middleware.Spec.Errors == nil {
		return nil
	}

	return &model.Middleware{
		BaseResource: model.BaseResource{
			Name:      middleware.Name,
			Namespace: middleware.Namespace,
			Labels:    convertK8sDataToModelData(middleware.Labels),
		},
		Type: model.MiddlewareTypeErrorPage,
		ErrorPage: &model.MiddlewareErrorPage{
			Status: middleware.Spec.Errors.Status,
			Service: &model.Service{
				BaseResource: model.BaseResource{
					Name:      middleware.Spec.Errors.Service.Name,
					Namespace: middleware.Spec.Errors.Service.Namespace,
				},
				Port: int32(middleware.Spec.Errors.Service.Port.IntValue()),
			},
			Query: middleware.Spec.Errors.Query,
		},
	}
}

// todo: digestAuth

func (k K8sOrchestratedServiceManager) CreateNewErrorPageMiddleware(ctx context.Context, namespace, name string, status []string, service *model.Service, query string, labels []model.KeyValue) (*model.Middleware, error) {
	k8sLables := convertModelDataToK8sData(labels)

	middleware, err := k.traefikClient.
		TraefikV1alpha1().
		Middlewares(namespace).
		Create(ctx,
			&v1alpha1.Middleware{
				ObjectMeta: v1.ObjectMeta{
					Name:      name,
					Labels:    k8sLables,
					Namespace: namespace,
				},
				Spec: v1alpha1.MiddlewareSpec{
					Errors: &v1alpha1.ErrorPage{
						Status: status,
						Service: v1alpha1.Service{
							LoadBalancerSpec: v1alpha1.LoadBalancerSpec{
								Name:      service.Name,
								Port:      intstr.FromInt(int(service.Port)),
								Kind:      "Service",
								Namespace: service.Namespace,
							},
						},
						Query: query,
					},
				},
			},
			v1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating error page middleware: %v", err)
	}
	middlewareModel := convertK8sMiddlewareToModelMiddleware(middleware)
	if middlewareModel == nil {
		return nil, fmt.Errorf("middleware not found")
	}
	return middlewareModel, nil
}

func (k K8sOrchestratedServiceManager) GetMiddleware(ctx context.Context, namespace, middlewareName string) (*model.Middleware, error) {
	middleware, err := k.traefikClient.
		TraefikV1alpha1().
		Middlewares(namespace).
		Get(ctx, middlewareName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting middleware: %v", err)
	}
	middlewareModel := convertK8sMiddlewareToModelMiddleware(middleware)
	if middlewareModel == nil {
		return nil, fmt.Errorf("middleware not found")
	}
	return middlewareModel, nil
}
