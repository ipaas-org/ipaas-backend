package k8smanager

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

func convertK8sDeploymentToModelDeployment(deployment *appsv1.Deployment) *model.Deployment {
	return &model.Deployment{
		BaseResource: model.BaseResource{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			Labels:    convertK8sDataToModelData(deployment.Labels),
		},
		Replicas:      *deployment.Spec.Replicas,
		ImageRegistry: deployment.Spec.Template.Spec.Containers[0].Image,
		CpuLimits:     deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String(),
		MemoryLimits:  deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String(),
		Port:          deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort,
	}
}

func (k K8sOrchestratedServiceManager) GetDeployment(ctx context.Context, namespace, deploymentName string) (*model.Deployment, error) {
	deployment, err := k.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting deployment: %v", err)
	}

	return convertK8sDeploymentToModelDeployment(deployment), nil
}

func (k K8sOrchestratedServiceManager) CreateNewDeployment(ctx context.Context, namespace, deploymentName, app, imageRegistry string, replicas, port int32, labels []model.KeyValue, configMapName string, volume *model.Volume) (*model.Deployment, error) {
	k8sLabels := convertModelDataToK8sData(labels)
	if k8sLabels[model.AppLabel] == "" {
		k8sLabels[model.AppLabel] = app
	}

	deployment := &appsv1.Deployment{
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
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: k8sLabels,
				},
				Spec: corev1.PodSpec{
					//to allow image updates on rollouts
					RestartPolicy: corev1.RestartPolicyAlways,
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "registrypullsecret",
						}},
					Containers: []corev1.Container{
						{
							Name:  app,
							Image: imageRegistry,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    k.cpuResource,
									corev1.ResourceMemory: k.memoryResource,
								}},
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: configMapName,
										},
									}}},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port,
								}},
						},
					},
				},
			},
		}}
	if volume != nil {
		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: volume.Name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: volume.PersistantVolumeClaim.Name,
					},
				},
			}}
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      volume.Name,
				MountPath: volume.MountPath,
			},
		}
	}

	createdDeployment, err := k.clientset.AppsV1().Deployments(namespace).
		Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating deployment: %v", err)
	}

	return convertK8sDeploymentToModelDeployment(createdDeployment), nil
}

func (k K8sOrchestratedServiceManager) UpdateDeployment(ctx context.Context, namespace, deploymentName string, imageRegistry string, replicas, port int32, labels []model.KeyValue, configMapName string) (*model.Deployment, error) {
	deployment, err := k.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting deployment: %v", err)
	}

	deployment.Spec.Replicas = &replicas
	deployment.Spec.Template.Spec.Containers[0].Image = imageRegistry
	deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = port
	if configMapName != "" {
		deployment.Spec.Template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
			{
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
				},
			}}
	}

	k8sLabels := convertModelDataToK8sData(labels)
	deployment.Labels = k8sLabels
	deployment.Spec.Template.Labels = k8sLabels

	updatedDeployment, err := k.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error updating deployment: %v", err)
	}

	return convertK8sDeploymentToModelDeployment(updatedDeployment), nil
}

func (k K8sOrchestratedServiceManager) DeleteDeployment(ctx context.Context, namespace, deploymentName string, gracePeriod int64) error {
	grace := &gracePeriod
	if gracePeriod < 0 {
		grace = nil
	}
	err := k.clientset.AppsV1().Deployments(namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{
		GracePeriodSeconds: grace,
	})
	if err != nil {
		return fmt.Errorf("error deleting deployment: %v", err)
	}
	return nil
}

func (k K8sOrchestratedServiceManager) RestartDeployment(ctx context.Context, namespace string, deploymentName string) error {
	_, err := k.clientset.AppsV1().Deployments(namespace).Patch(ctx,
		deploymentName, types.MergePatchType, []byte(`{"spec":{"template":{"metadata":{"annotations":{"date":"`+metav1.Now().String()+`"}}}}}`), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error restarting deployment: %v", err)
	}
	return nil
}

func (k K8sOrchestratedServiceManager) WaitDeploymentReadyState(ctx context.Context, namespace, deploymentName string) (chan struct{}, chan error) {
	done := make(chan struct{})
	errChan := make(chan error)
	timeout := int64(5 * 60)
	deploymentWatch, err := k.clientset.AppsV1().Deployments(namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector:  model.ResourceNameLabel + "=" + deploymentName,
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
				if e.Type != watch.Modified {
					continue
				}
				deployment, ok := e.Object.(*appsv1.Deployment)
				if !ok {
					//todo: chose what to do if this happens
					errChan <- fmt.Errorf("error casting object to deployment")
				}
				for _, condition := range deployment.Status.Conditions {
					if condition.Type == appsv1.DeploymentAvailable &&
						condition.Status == corev1.ConditionTrue &&
						condition.Reason == "MinimumReplicasAvailable" {
						done <- struct{}{}
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return done, errChan
}
