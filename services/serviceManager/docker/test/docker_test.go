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
	envs := []model.Env{
		{Key: "key1", Value: "value1"},
		{Key: "vano", Value: "vano"},
		{Key: "test", Value: "test"},
	}

	id, name, err := containerManager.CreateNewContainer(context.Background(), image, envs)
	if err != nil {
		t.Errorf("error creating the container: %v", err)
	}

	err = containerManager.StartContainer(context.Background(), id)
	if err != nil {
		t.Error(err)
	}
	t.Logf("container id %s and name %s", id, name)
	// t.Cleanup(func() {
	// 	//remove the container with id
	// 	err := containerManager.RemoveContainer(context.Background(), id)
	// 	if err != nil {
	// 		t.Errorf("error removing the container: %v", err)
	// 	}
	// })
}
