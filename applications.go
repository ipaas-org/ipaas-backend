package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/ipaas-org/ipaas-backend/model"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateNewApplicationFromRepo creates a container from an image which is the one created from a student's repository
func (c ContainerController) CreateNewApplicationFromRepo(creatorID int, port, name, language, imageName string) (string, error) {
	//generic configs for the container
	containerConfig := &container.Config{
		Image: imageName,
	}

	// externalPort, err := getFreePort()
	// if err != nil {
	// 	return "", err
	// }

	//host bindings config, hostPort is not set cause the engine will assign a dinamyc one
	hostBinding := nat.PortBinding{
		HostIP: "0.0.0.0",
		//HostPort is the port that the host will listen to, since it's not set
		//the docker engine will assign a random open port
		// HostPort: strconv.Itoa(externalPort),
	}

	//set the port for the container (internal one)
	containerPort, err := nat.NewPort("tcp", port)
	if err != nil {
		return "", err
	}

	//set a slice of possible port bindings
	//since it's a db container we need just one
	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}

	//set the configuration of the host
	//set the port bindings and the restart policy
	//!choose a restart policy
	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
		// RestartPolicy: container.RestartPolicy{
		// 	Name:              "on-failure",
		// 	MaximumRetryCount: 3,
		// },
	}

	//create the container
	containerBody, err := c.cli.ContainerCreate(c.ctx, containerConfig,
		hostConfig, nil, nil, fmt.Sprintf("%d-%s-%s", creatorID, name, language))
	if err != nil {
		return "", err
	}

	return containerBody.ID, nil
}

// GetAppInfoFromContainer returns the application metadata from the container id.
// The parameter checkCommit will make the function check if the last commit changed, if so the application
// returned will have isUpdatable set to true.
func (c ContainerController) GetAppInfoFromContainer(containerId string, checkCommit bool, util *Util) (model.Application, error) {
	db, err := connectToDB()
	if err != nil {
		return model.Application{}, err
	}
	defer db.Client().Disconnect(context.TODO())

	var app model.Application
	err = db.Collection("applications").FindOne(context.TODO(), bson.M{"containerID": containerId}).Decode(&app)
	if err != nil {
		return model.Application{}, err
	}

	if checkCommit {
		app.IsUpdatable, err = util.HasLastCommitChanged(app.LastCommitHash, app.GithubRepo, app.GithubBranch)
		if err != nil {
			return model.Application{}, err
		}
	}

	// if getEnvs {
	// 	envs := make(map[string]string)
	// 	results, err := db.Query("SELECT * FROM envs WHERE applicationID=?", app.ID)
	// 	if err != nil {
	// 		return Application{}, nil, err
	// 	}
	// 	for results.Next() {
	// 		var key, value string
	// 		err = results.Scan(&key, &value)
	// 		if err != nil {
	// 			return Application{}, nil, err
	// 		}
	// 		envs[key] = value
	// 	}
	// 	return app, envs, nil
	// }

	return app, nil
}
