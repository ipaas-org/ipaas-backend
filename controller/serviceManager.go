package controller

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ipaas-org/ipaas-backend/model"
)

func (c *Controller) getDefaultLabels(owner, environment, app, appID, resourceName string) []model.KeyValue {
	labels := []model.KeyValue{
		{Key: model.OwnerLabel, Value: owner},
		{Key: model.EnvironmentLabel, Value: staticTempEnvironment},
		{Key: model.IpaasVersionLabel, Value: c.config.Version},
		{Key: model.IpaasManagedLabel, Value: "true"},
		{Key: model.ResourceNameLabel, Value: resourceName},
	}
	if app != "" {
		labels = append(labels, []model.KeyValue{
			{Key: model.AppLabel, Value: app},
			{Key: model.AppIDLabel, Value: appID},
		}...)
	}
	return labels
}

func (c *Controller) createConfigMap(ctx context.Context, app *model.Application, user *model.User, envs []model.KeyValue) (*model.ConfigMap, error) {
	c.l.Debugf("creating config map for application %s", app.Name)
	resourceName := fmt.Sprintf("cm-%s", app.Name)
	configMapLabels := c.getDefaultLabels(user.Code, staticTempEnvironment, app.Name, app.ID.Hex(), resourceName)
	configMap, err := c.ServiceManager.CreateNewConfigMap(ctx, user.Namespace, resourceName, envs, configMapLabels)
	if err != nil {
		c.l.Errorf("error creating config map: %v", err)
		return nil, err
	}
	c.l.Debugf("created config map for %s in namespace: %s with name: %s", app.Name, configMap.Namespace, configMap.Name)
	return configMap, nil
}

func (c *Controller) createDeployment(ctx context.Context, app *model.Application, user *model.User, registryImage, configMapName string, volume *model.Volume) (*model.Deployment, error) {
	c.l.Debugf("creating deployment for application %s", app.Name)
	resourceName := fmt.Sprintf("deploy-%s", app.Name)
	deploymentLabels := c.getDefaultLabels(user.Code, staticTempEnvironment, app.Name, app.ID.Hex(), resourceName)
	deploymentLabels = append(deploymentLabels,
		[]model.KeyValue{
			{
				Key:   model.VisibilityLabel,
				Value: app.Visiblity,
			}, {
				Key:   model.PortLabel,
				Value: app.ListeningPort,
			}}...)
	p, err := strconv.Atoi(app.ListeningPort)
	if err != nil {
		c.l.Errorf("error converting port to int: %v", err)
		return nil, err
	}
	intPort := int32(p)
	deployment, err := c.ServiceManager.CreateNewDeployment(ctx, user.Namespace, resourceName, app.Name, registryImage, 1, intPort, deploymentLabels, configMapName, volume)
	if err != nil {
		c.l.Errorf("error creating deployment: %v", err)
		return nil, err
	}
	c.l.Debugf("created deployment for %s in namespace: %s with name: %s", app.Name, deployment.Namespace, deployment.Name)
	return deployment, nil
}

func (c *Controller) createService(ctx context.Context, app *model.Application, user *model.User) (*model.Service, error) {
	c.l.Debugf("creating service for application %s", app.Name)
	// resourceName := fmt.Sprintf("svc-%s", app.Name)
	resourceName := app.Name
	serviceLabels := c.getDefaultLabels(user.Code, staticTempEnvironment, app.Name, app.ID.Hex(), resourceName)
	p, err := strconv.Atoi(app.ListeningPort)
	if err != nil {
		c.l.Errorf("error converting port to int: %v", err)
		return nil, err
	}
	intPort := int32(p)
	service, err := c.ServiceManager.CreateNewService(ctx, user.Namespace, resourceName, app.Name, intPort, serviceLabels)
	if err != nil {
		c.l.Errorf("error creating service: %v", err)
		return nil, err
	}
	c.l.Debugf("created service for %s in namespace: %s with name: %s", app.Name, service.Namespace, service.Name)
	return service, nil
}

func (c *Controller) createIngressRoute(ctx context.Context, app *model.Application, user *model.User, host, serviceName string, listeningPort int32) (*model.IngressRoute, error) {
	c.l.Debugf("creating ingress route for application %s", app.Name)
	resourceName := fmt.Sprintf("ir-%s", app.Name)
	ingressRouteLabels := c.getDefaultLabels(user.Code, staticTempEnvironment, app.Name, app.ID.Hex(), resourceName)
	// host := fmt.Sprintf("%s.%s", app.Name, c.app.BaseDefaultDomain)
	ingressRoute, err := c.ServiceManager.CreateNewIngressRoute(ctx, user.Namespace, resourceName, host, serviceName, listeningPort, ingressRouteLabels)
	if err != nil {
		c.l.Errorf("error creating ingress route: %v", err)
		return nil, err
	}
	c.l.Debugf("created ingress route for %s in namespace: %s with name: %s", app.Name, ingressRoute.Namespace, ingressRoute.Name)
	return ingressRoute, nil
}

func (c *Controller) CreatePersistantVolumeClaim(ctx context.Context, app *model.Application, user *model.User, storageClass string, GiSize int64) (*model.PersistentVolumeClaim, error) {
	c.l.Debugf("creating persistant volume claim for application %s", app.Name)
	pvcName := fmt.Sprintf("pvc-%s", app.Name)
	pvcLabels := c.getDefaultLabels(user.Code, staticTempEnvironment, app.Name, app.ID.Hex(), pvcName)
	pvc, err := c.ServiceManager.CreateNewPersistentVolumeClaim(ctx, user.Namespace, pvcName, storageClass, GiSize, pvcLabels)
	if err != nil {
		c.l.Errorf("error creating persistant volume claim: %v", err)
		return nil, err
	}
	c.l.Debugf("created persistant volume claim for %s in namespace: %s with name: %s", app.Name, pvc.Namespace, pvc.Name)
	return pvc, nil
}
