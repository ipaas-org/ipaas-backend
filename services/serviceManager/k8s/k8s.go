package k8smanager

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	traefikv "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sOrchestratedServiceManager struct {
	clientset      kubernetes.Interface
	kubeConfig     *rest.Config
	traefikClient  *traefikv.Clientset
	cpuResource    resource.Quantity
	memoryResource resource.Quantity
}

func NewK8sOrchestratedServiceManager(kubeConfigPath, cpuResource, memoryResource string) (*K8sOrchestratedServiceManager, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes config: %v", err)
	}

	traefikClient, err := traefikv.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating traefik client: %v", err)
	}

	cpuQuantity, err := resource.ParseQuantity(cpuResource)
	if err != nil {
		return nil, fmt.Errorf("error parsing cpu resource: %v", err)
	}

	memoryQuantity, err := resource.ParseQuantity(memoryResource)
	if err != nil {
		return nil, fmt.Errorf("error parsing memory resource: %v", err)
	}
	return &K8sOrchestratedServiceManager{
		clientset:      clientset,
		kubeConfig:     kubeConfig,
		traefikClient:  traefikClient,
		cpuResource:    cpuQuantity,
		memoryResource: memoryQuantity,
	}, nil
}

func (k K8sOrchestratedServiceManager) CreateNewNamespace(ctx context.Context, namespace string, labels []model.KeyValue) error {
	k8sLabels := convertModelKeyValuesToLables(labels)
	_, err := k.clientset.CoreV1().Namespaces().Create(ctx,
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   namespace,
				Labels: k8sLabels,
				// Labels: map[string]string{
				// 	ownerLabel:        owner,
				// 	environmentLabel:  environment,
				// 	ipaasVersionLabel: k.appVersion,
				// 	ipaasManagedLabel: "true",
				// },
			},
		}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating namespace: %v", err)
	}
	return nil
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

func (k K8sOrchestratedServiceManager) CreateDeployment(ctx context.Context, namespace, deploymentName, app, imageRegistry string, replicas, port int32, labels []model.KeyValue, configMap *model.ConfigMap) (*model.Deployment, error) {
	// labels := map[string]string{
	// 	appLabel:          app,
	// 	ownerLabel:        owner,
	// 	environmentLabel:  environment,
	// 	visibilityLabel:   visibility,
	// 	portLabel:         fmt.Sprintf("%d", port),
	// 	ipaasVersionLabel: k.appVersion,
	// 	ipaasManagedLabel: "true",
	// }
	k8sLabels := convertModelKeyValuesToLables(labels)
	if k8sLabels[model.AppLabel] == "" {
		k8sLabels[model.AppLabel] = app
	}
	_, err := k.clientset.AppsV1().Deployments(namespace).
		Create(ctx,
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:   deploymentName,
					Labels: k8sLabels,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							model.AppLabel: app,
						},
					},
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: k8sLabels,
						},
						Spec: v1.PodSpec{
							//to allow image updates on rollouts
							RestartPolicy: v1.RestartPolicyAlways,
							ImagePullSecrets: []v1.LocalObjectReference{
								{
									Name: "registrypullsecret",
								},
							},
							Containers: []v1.Container{
								{
									Name:  app,
									Image: imageRegistry,
									Resources: v1.ResourceRequirements{
										Limits: v1.ResourceList{
											v1.ResourceCPU:    k.cpuResource,
											v1.ResourceMemory: k.memoryResource,
										},
									},
									EnvFrom: []v1.EnvFromSource{
										{
											ConfigMapRef: &v1.ConfigMapEnvSource{
												LocalObjectReference: v1.LocalObjectReference{
													Name: configMap.Name,
												},
											},
										},
									},
									Ports: []v1.ContainerPort{
										{
											ContainerPort: port,
										},
									},
								},
							},
						},
					},
				}}, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating deployment: %v", err)
	}
	return &model.Deployment{
		BaseResource: model.BaseResource{
			Name:      deploymentName,
			Namespace: namespace,
			Labels:    labels,
		},
		Replicas:      replicas,
		ImageRegistry: imageRegistry,
		CpuLimits:     k.cpuResource.String(),
		MemoryLimits:  k.memoryResource.String(),
		// Envs:          envs,
		Port: port,
	}, nil
}

func (k K8sOrchestratedServiceManager) RestartDeployment(ctx context.Context, namespace string, deploymentName string) error {
	_, err := k.clientset.AppsV1().Deployments(namespace).Patch(ctx,
		deploymentName, types.MergePatchType, []byte(`{"spec":{"template":{"metadata":{"annotations":{"date":"`+metav1.Now().String()+`"}}}}}`), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error restarting deployment: %v", err)
	}
	return nil
}

func (k K8sOrchestratedServiceManager) CreateService(ctx context.Context, namespace, serviceName, app string, port int32, labels []model.KeyValue) (*model.Service, error) {
	// labels := map[string]string{
	// 	appLabel:          app,
	// 	ownerLabel:        owner,
	// 	environmentLabel:  environment,
	// 	visibilityLabel:   visibility,
	// 	ipaasManagedLabel: "true",
	// }
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
							Port: 80,
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
		Port:       80,
		TargetPort: port,
	}, nil
}

func (k K8sOrchestratedServiceManager) CreateIngressRoute(ctx context.Context, namespace, ingressRouteName, host, serviceName string, labels []model.KeyValue) (*model.IngressRoute, error) {
	// labels := map[string]string{
	// 	appLabel:          app,
	// 	ownerLabel:        owner,
	// 	environmentLabel:  environment,
	// 	visibilityLabel:   visibility,
	// 	ipaasManagedLabel: "true",
	// }
	k8sLables := convertModelKeyValuesToLables(labels)
	entrpoints := []string{"web", "websecure"}
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
					EntryPoints: entrpoints,
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
											IntVal: 80,
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
		Entrypoints: entrpoints,
		Domain:      host,
	}, nil
}

func (k K8sOrchestratedServiceManager) CreateConfigMap(ctx context.Context, namespace, configMapName string, data, labels []model.KeyValue) (*model.ConfigMap, error) {
	// labels := map[string]string{
	// 	appLabel:          app,
	// 	ownerLabel:        owner,
	// 	environmentLabel:  environment,
	// 	ipaasVersionLabel: k.appVersion,
	// 	ipaasManagedLabel: "true",
	// }

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

func (k K8sOrchestratedServiceManager) UpdateConfigMap(ctx context.Context, namespace, configMapName string, data []model.KeyValue) (*model.ConfigMap, error) {
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

	return &model.ConfigMap{
		BaseResource: model.BaseResource{
			Name:      configMapName,
			Namespace: namespace,
			// Labels:    labels,
		},
	}, nil
}
