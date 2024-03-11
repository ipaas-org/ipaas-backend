package controller

import (
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
)

func (c *Controller) GenerateLabels(name, ownerID string, serviceKind model.ApplicationKind) []model.KeyValue {
	return []model.KeyValue{
		{Key: "org.ipaas.service.name", Value: name},
		{Key: "org.ipaas.service.owner", Value: ownerID},
		{Key: "org.ipaas.service.kind", Value: string(serviceKind)},
		{Key: "org.ipaas.version", Value: c.app.Version},
		{Key: "org.ipaas.name", Value: c.app.Name},
		{Key: "org.ipaas.deployment", Value: c.app.Deployment},
	}
}

func (c *Controller) GenerateTraefikDnsLables(name, host, port string) []model.KeyValue {
	router := fmt.Sprintf("traefik.http.routers.%s", name)
	service := fmt.Sprintf("traefik.http.services.%s", name)

	return []model.KeyValue{
		{Key: "traefik.enable", Value: "true"},
		{Key: router + ".entrypoints", Value: "web"},
		{Key: router + ".rule", Value: "Host(`" + host + "`)"},
		{Key: service + ".loadbalancer.server.port", Value: port},
	}
}

// func (c *Controller) createConnectAndStartContainer(ctx context.Context, name, imageID, networkID string, envs, labels []model.KeyValue) (*model.Container, error) {
// 	c.l.Debugf("creating new container with name: %s, image: %s, envs: %v, labels: %v", name, imageID, envs, labels)
// 	container, err := c.serviceManager.CreateNewService(ctx, name, imageID, envs, labels)
// 	if err != nil {
// 		return nil, err
// 	}
// 	//connect container to user's network id and set as dns the application name
// 	if err := c.serviceManager.ConnectServiceToNetwork(ctx, container.ID, networkID, name); err != nil {
// 		return nil, err
// 	}

// 	if err := c.serviceManager.StartServiceByID(ctx, container.ID); err != nil {
// 		return nil, err
// 	}

// 	return container, nil
// }

// func (c *Controller) CreateNewContainer(ctx context.Context, kind model.ServiceKind, ownerID, name, image string, env, labels []model.KeyValue) (*model.Container, error) {
// 	return c.serviceManager.CreateNewContainer(ctx, name, image, env, labels)
// }

// func (c *Controller) RemoveContainer(ctx context.Context, id string) error {
// 	return c.serviceManager.RemoveContainerByID(ctx, id)
// }

// func (c *Controller) StartContainer(ctx context.Context, id string) error {
// 	return c.serviceManager.StartContainerByID(ctx, id)
// }

// func (c *Controller) CreateContainerFromIDAndImage(ctx context.Context, id, lastCommitHash, image string) error {
// 	uuid, err := primitive.ObjectIDFromHex(id)
// 	if err != nil {
// 		c.l.Errorf("error parsing uuid: %v", err)
// 		return fmt.Errorf("primitive.ObjectIDFromHex: %w", err)
// 	}

// 	app, err := c.GetApplicationByID(ctx, uuid)
// 	if err != nil {
// 		c.l.Errorf("error finding application: %v", err)
// 		return fmt.Errorf("c.applicationRepo.FindByID: %w", err)
// 	}

// 	labels := c.GenerateLabels(app.Name, app.Owner, model.ApplicationKindWeb)

// 	c.l.Debugf("creating new container for application: %+v", app)
// 	container, err := c.CreateNewContainer(ctx, model.ApplicationKindWeb, app.Owner, app.Name, image, app.Envs, labels)
// 	if err != nil {
// 		c.l.Errorf("error creating new container: %v", err)
// 		return fmt.Errorf("c.CreateNewContainer: %w", err)
// 	}

// 	app.State = StateCreated
// 	app.Container = container
// 	app.Kind = model.ApplicationKindWeb
// 	app.LastCommitHash = lastCommitHash

// 	c.l.Debugf("container created correctly, updating application status: %+v", app)
// 	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
// 		c.l.Errorf("error updating application status: %v", err)
// 		return fmt.Errorf("c.applicationRepo.UpdateByID: %w", err)
// 	}
// 	c.l.Debug("application status updated correctly")
// 	c.l.Debug("starting container")
// 	if err := c.StartContainer(ctx, container.ContainerID); err != nil {
// 		app.State = StateFailed
// 		if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
// 			c.l.Errorf("error updating container: %v", err)
// 		}
// 		c.l.Errorf("error starting container: %v", err)
// 		return fmt.Errorf("c.StartContainer: %w", err)
// 	}

// 	return nil
// }

//=========== network ===========

// func (c *Controller) RemoveNetwork(ctx context.Context, id string) error {
// 	return c.serviceManager.RemoveNetwork(ctx, id)
// }

// func (c *Controller) CreateNewNetwork(ctx context.Context, name string) (string, error) {
// 	return c.serviceManager.CreateNewNetwork(ctx, name)
// }

// func (c *Controller) ConnectContainerToNetwork(ctx context.Context, containerID, networkID, dnsAlias string) error {
// 	return c.serviceManager.ConnectContainerToNetwork(ctx, containerID, networkID, dnsAlias)
// }
