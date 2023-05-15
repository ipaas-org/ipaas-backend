package docker

import (
	"context"
	"testing"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/services/serviceManager/docker"
)

func TestCreateNewContainer(t *testing.T) {
	containerManager, err := docker.NewDockerApplicationManager()
	if err != nil {
		t.Fatalf("error creating containerManager: %v", err)
	}

	image := "traefik/whoami:latest"
	envs := []model.KeyValue{
		{Key: "key1", Value: "value1"},
		{Key: "vano", Value: "vano"},
		{Key: "test", Value: "test"},
	}

	labels := []model.KeyValue{
		{Key: "org.ipaas.service.type", Value: "test"},
	}

	networkID := "65e0226c67d8e04ef12dd1e046c7d25b0e26db131d028dfaa83852120319ebf3"
	dnsAlias := "test"
	ctx := context.Background()
	id, name, err := containerManager.CreateNewContainer(ctx, "test", image, envs, labels)
	if err != nil {
		t.Errorf("error creating the container: %v", err)
	}

	if err := containerManager.ConnectContainerToNetwork(ctx, id, networkID, dnsAlias); err != nil {
		t.Errorf("error connecting container to network: %v", err)
	}

	err = containerManager.StartContainer(ctx, id)
	if err != nil {
		t.Error(err)
	}
	t.Logf("container id %s and name %s", id, name)
	// t.Cleanup(func() {
	// 	//remove the container with id
	// 	err := containerManager.RemoveContainer(ctx, id)
	// 	if err != nil {
	// 		t.Errorf("error removing the container: %v", err)
	// 	}
	// })
}
