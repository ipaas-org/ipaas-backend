package events

import (
	"context"
	"runtime/debug"

	"github.com/docker/docker/client"
	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/sirupsen/logrus"
)

type ContainerEventHandler struct {
	cli        *client.Client
	controller *controller.Controller
	l          *logrus.Logger
	Done       chan struct{}
}

func NewContainerEventHandler(ctx context.Context, controller *controller.Controller, l *logrus.Logger) (*ContainerEventHandler, error) {
	c := new(ContainerEventHandler)
	var err error
	//creating docker client from env

	c.cli, err = client.NewClientWithOpts(client.FromEnv)
	c.cli.NegotiateAPIVersion(ctx)
	if err != nil {
		return nil, err
	}

	c.controller = controller
	c.l = l
	c.Done = make(chan struct{})

	return c, nil
}

func StartConainerEventHandler(ctx context.Context, c *ContainerEventHandler, ID int, RoutineMonitor chan int) {
	defer func() {
		if r := recover(); r != nil {
			c.l.Errorf("router panic, recovering: \nerror: %v\n\nstack: %s", r, string(debug.Stack()))
		}
		if ctx.Err() != nil {
			RoutineMonitor <- ID
		} else {
			c.l.Info("ContainerEventHandler not restarting, context was canceled")
			c.Done <- struct{}{}
		}
	}()

	c.l.Info(">>> Starting ContainerEventHandler >>>")

	controller := c.controller
	logger := c.l
	var err error
	c, err = NewContainerEventHandler(ctx, controller, logger)
	if err != nil {
		logger.Errorf("error creating ContainerEventHandler: %v", err)
		return
	}
	c.l.Warn("Event handler not implemented")
	// c.start(ctx)
}

// func (c *ContainerEventHandler) start(ctx context.Context) {
// 	eventChan, errChan := c.cli.Events(ctx, types.EventsOptions{})
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			c.l.Info("ContainerEventHandler context canceled, stopping")
// 			return

// 		case event := <-eventChan:
// 			//log.Println(event)
// 			switch event.Type {
// 			case "container":
// 				//log.Println("container event\n\n")
// 				switch event.Action {
// 				case "die", "kill":
// 					c.l.Infof("[EVENT] Container %s: %s", event.Action, event.Actor.ID)
// 					// app, err := c.controller.ApplicationRepo.FindByName(ctx, event.Actor.ID)
// 					if err != nil {
// 						c.l.Infof("a container died or was killed but it's not handled by ipaas: %v", err)
// 						continue
// 					}
// 					if app.State == model.ApplicationStateDeleting {
// 						c.l.Info("application is being deleted, passing")
// 						continue
// 					}
// 					c.l.Info("unexpected container death, notifying user")
// 					app.State = model.ApplicationStateCrashed
// 					if _, err := c.controller.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
// 						c.l.Errorf("error updating application: %v", err)
// 						continue
// 					}
// 				//todo: notify user
// 				// case "health_status":
// 				// 	log.Println("[EVENT] Container health status:", event.Actor.ID)
// 				case "start":
// 					c.l.Infof("[EVENT] Container %s: %s", event.Action, event.Actor.ID)
// 					// app, err := c.controller.ApplicationRepo.FindByContainerID(ctx, event.Actor.ID)
// 					if err != nil {
// 						c.l.Infof("a container started but it's not handled by ipaas: %v", err)
// 						continue
// 					}
// 					if app.State != model.ApplicationStateCrashed {
// 						c.l.Info("application is not crashed, passing")
// 						continue
// 					}
// 					c.l.Info("container started, updating state")
// 					app.State = model.ApplicationStateRunning
// 					if _, err := c.controller.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
// 						c.l.Errorf("error updating application: %v", err)
// 						continue
// 					}

// 					// default:
// 					// 	c.l.Infof("[EVENT] Container %s: %s", event.Action, event.Actor.ID)
// 				}
// 			}
// 		case err := <-errChan:
// 			c.l.Errorf("error from docker event stream: %v", err)

// 		}
// 	}
// }
