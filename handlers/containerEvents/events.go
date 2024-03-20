package events

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type watchGeneratorFunc func(context.Context) (watch.Interface, error)

type ContainerEventHandler struct {
	// cli        *client.Client
	watch          watch.Interface
	watchGenerator watchGeneratorFunc
	controller     *controller.Controller
	l              *logrus.Logger
	Done           chan struct{}
}

func NewContainerEventHandler(ctx context.Context, controller *controller.Controller, watchGenerator watchGeneratorFunc, l *logrus.Logger) (*ContainerEventHandler, error) {
	c := new(ContainerEventHandler)
	c.controller = controller
	c.l = l
	c.Done = make(chan struct{})
	c.watchGenerator = watchGenerator
	w, err := c.watchGenerator(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting watch: %v", err)
	}
	c.watch = w
	return c, nil
}

func StartContainerEventHandler(ctx context.Context, c *ContainerEventHandler, ID int, RoutineMonitor chan int) {
	defer func() {
		if c.watch != nil {
			c.watch.Stop()
		}
		if r := recover(); r != nil {
			c.l.Errorf("container event handler panic, recovering: \nerror: %v\n\nstack: %s", r, string(debug.Stack()))
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
	c, err = NewContainerEventHandler(ctx, controller, c.watchGenerator, logger)
	if err != nil {
		logger.Errorf("error creating ContainerEventHandler: %v", err)
		return
	}
	c.start(ctx)
}

func (c *ContainerEventHandler) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.l.Info("ContainerEventHandler context canceled, stopping")
			return

		case event := <-c.watch.ResultChan():
			//log.Println(event)
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			f := make(logrus.Fields)
			for _, c := range pod.Status.Conditions {
				var value interface{}
				if c.Status == "True" {
					value = true
				} else {
					value = c.Reason
				}
				f["condition-"+strings.ToLower(string(c.Type))] = value
			}
			c.l.WithFields(f).Debugf("[EVENT] pod %s", pod.Name)

			appID := pod.Labels[model.AppIDLabel]
			appPrimitiveID, err := primitive.ObjectIDFromHex(appID)
			if err != nil {
				c.l.Errorf("error parsing app id to object id: %v", err)
				continue
			}
			app, err := c.controller.GetApplicationByID(ctx, appPrimitiveID)
			if err != nil {
				c.l.Errorf("error getting application by id: %v", err)
				continue
			}

			if app.State == model.ApplicationStateDeleting {
				c.l.Info("application is being deleted, passing")
				continue
			}
			// switch event.Type {
			// case "container":
			// 	//log.Println("container event\n\n")
			// 	switch event.Action {
			// 	case "die", "kill":
			// 		c.l.Infof("[EVENT] Container %s: %s", event.Action, event.Actor.ID)
			// 		// app, err := c.controller.ApplicationRepo.FindByName(ctx, event.Actor.ID)
			// 		if err != nil {
			// 			c.l.Infof("a container died or was killed but it's not handled by ipaas: %v", err)
			// 			continue
			// 		}
			// 		if app.State == model.ApplicationStateDeleting {
			// 			c.l.Info("application is being deleted, passing")
			// 			continue
			// 		}
			// 		c.l.Info("unexpected container death, notifying user")
			// 		app.State = model.ApplicationStateCrashed
			// 		if _, err := c.controller.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
			// 			c.l.Errorf("error updating application: %v", err)
			// 			continue
			// 		}
			// 	//todo: notify user
			// 	// case "health_status":
			// 	// 	log.Println("[EVENT] Container health status:", event.Actor.ID)
			// 	case "start":
			// 		c.l.Infof("[EVENT] Container %s: %s", event.Action, event.Actor.ID)
			// 		// app, err := c.controller.ApplicationRepo.FindByContainerID(ctx, event.Actor.ID)
			// 		if err != nil {
			// 			c.l.Infof("a container started but it's not handled by ipaas: %v", err)
			// 			continue
			// 		}
			// 		if app.State != model.ApplicationStateCrashed {
			// 			c.l.Info("application is not crashed, passing")
			// 			continue
			// 		}
			// 		c.l.Info("container started, updating state")
			// 		app.State = model.ApplicationStateRunning
			// 		if _, err := c.controller.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
			// 			c.l.Errorf("error updating application: %v", err)
			// 			continue
			// 		}

			// 		// default:
			// 		// 	c.l.Infof("[EVENT] Container %s: %s", event.Action, event.Actor.ID)
			// 	}
			// }
		}
	}
}
