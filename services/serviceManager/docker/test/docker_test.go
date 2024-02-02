package docker

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/services/serviceManager/docker"
)

func TestCreateNewContainer(t *testing.T) {
	ctx := context.Background()

	containerManager, err := docker.NewDockerApplicationManager(ctx)
	if err != nil {
		t.Fatalf("error creating containerManager: %v", err)
	}

	image := "busybox:latest"
	envs := []model.KeyValue{
		{Key: "key1", Value: "value1"},
		{Key: "vano", Value: "vano"},
	}

	labels := []model.KeyValue{
		{Key: "org.ipaas.service.type", Value: "test"},
	}

	rand := uuid.New().String()
	name := "container-test-" + rand

	container, err := containerManager.CreateNewService(ctx, name, image, envs, labels)
	if err != nil {
		t.Fatalf("error creating the container: %v", err)
	}

	t.Logf("container id %s and name %s", container.ID, container.Name)
	//remove the container with id
	err = containerManager.RemoveServiceByID(ctx, container.ID, true)
	if err != nil {
		t.Fatalf("error removing the container: %v", err)
	}
}

func TestStartContainer(t *testing.T) {
	ctx := context.Background()

	containerManager, err := docker.NewDockerApplicationManager(ctx)
	if err != nil {
		t.Fatalf("error creating containerManager: %v", err)
	}

	image := "busybox:latest"
	labels := []model.KeyValue{
		{Key: "org.ipaas.service.type", Value: "test"},
	}

	rand := uuid.New().String()
	name := "container-test-" + rand

	container, err := containerManager.CreateNewService(ctx, name, image, nil, labels)
	if err != nil {
		t.Fatalf("error creating the container: %v", err)
	}

	err = containerManager.StartServiceByID(ctx, container.ID)
	if err != nil {
		t.Fatalf("error starting the container: %v", err)
	}

	t.Logf("container id %s and name %s", container.ID, container.Name)
	//remove the container with id
	// err = containerManager.RemoveContainer(ctx, container)
	// if err != nil {
	// 	t.Errorf("error removing the container: %v", err)
	// }
}
