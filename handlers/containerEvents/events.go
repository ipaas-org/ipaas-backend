package events

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type watchGeneratorFunc func(context.Context) (watch.Interface, error)

var FakeWatchGeneratorFunc = func(ctx context.Context) (watch.Interface, error) {
	return nil, nil
}

type ContainerEventHandler struct {
	watch          watch.Interface
	watchGenerator watchGeneratorFunc
	controller     *controller.Controller
	l              *logrus.Logger
	Done           chan struct{}
}

func NewContainerEventHandler(ctx context.Context, controller *controller.Controller, watchGenerator watchGeneratorFunc, l *logrus.Logger) (*ContainerEventHandler, error) {
	l.Debug("creating container handler")
	c := new(ContainerEventHandler)
	c.controller = controller
	c.l = l
	c.Done = make(chan struct{})
	c.watchGenerator = watchGenerator
	w, err := c.watchGenerator(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting watch: %v", err)
	}
	l.Debug("created watch for container events")
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
		if ctx.Err() == nil {
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
			// c.l.Debugf("[EVENT] received %v event", event.Type)
			if event.Type == watch.Bookmark {
				c.l.Debugf("[EVENT] received bookmark event")
				c.l.Debugf("event: %+v", event)
			}
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
				f[fmt.Sprintf("condition%s", c.Type)] = value
			}
			c.l.WithFields(f).Debugf("[EVENT] pod %s", pod.Name)
			// c.l.Debugf("thing %+v", pod.Status.ContainerStatuses)
			if len(pod.Status.ContainerStatuses) == 0 {
				c.l.Infof("pod %s has no container statuses", pod.Name)
				continue
			}
			state := pod.Status.ContainerStatuses[0].State

			appID := pod.Labels[model.AppIDLabel]
			appPrimitiveID, err := primitive.ObjectIDFromHex(appID)
			if err != nil {
				c.l.Errorf("error parsing app id to object id: %v", err)
				continue
			}
			c.l.Debugf("app id: %s", appID)

			app, err := c.controller.GetApplicationByID(ctx, appPrimitiveID)
			if err != nil {
				c.l.Errorf("error getting application by id: %v", err)
				continue
			}

			c.l.Debugf("app %+v", app)
			// c.l.Debugf("event %+v", event)

			if state.Running != nil {
				if app.Service == nil || app.Service.Deployment == nil {
					c.l.Warnf("application %s has no service or deployment, probably currently being created, updating", app.Name)
					c.controller.AddCurrentPodToApplication(ctx, app.ID, pod.Name)
					continue
				}
				c.l.Info("currentPodName: ", app.Service.Deployment.CurrentPodName)
				if app.Service.Deployment.CurrentPodName != "" {
					if app.Service.Deployment.CurrentPodName != pod.Name {
						c.l.Warnf("application %s has a running container %s, but it's not the one we're watching %s, ignoring", app.Name, pod.Name, app.Service.Deployment.CurrentPodName)
						continue
					}
				} else {
					c.l.Debugf("application %s has a running container %s and we are watching none, watching", app.Name, pod.Name)
					app.Service.Deployment.CurrentPodName = pod.Name
					if _, err := c.controller.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
						c.l.Errorf("error updating application: %v", err)
					}
				}
				c.l.Debugf("state running: %+v\n", state.Running)
				c.l.Infof("container %s is running", pod.Name)
				if app.State != model.ApplicationStateRunning &&
					app.State != model.ApplicationStateDeleting {
					c.l.Infof("updating application state from %s to running", app.State)
					app.State = model.ApplicationStateRunning
					if _, err := c.controller.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
						c.l.Errorf("error updating application: %v", err)
						continue
					}
				}
				continue
			}

			if state.Terminated != nil {
				if app.State == model.ApplicationStateStarting {
					c.l.Infof("container %s is terminated, but application is still starting, there is a possible CrashLoopBackOff", pod.Name)
					app.State = model.ApplicationStateCrashed
					if _, err := c.controller.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
						c.l.Errorf("error updating application: %v", err)
					}
					continue
				}
				if app.Service == nil || app.Service.Deployment == nil {
					c.l.Warnf("application %s has no service or deployment, probably currently being created, ignoring", app.Name)
					continue
				}
				if app.Service.Deployment.CurrentPodName != pod.Name {
					c.l.Warnf("application %s has a terminated container %s, but it's not the one we're watching %s, ignoring", app.Name, pod.Name, app.Service.Deployment.CurrentPodName)
					continue
				}
				c.l.Debugf("state terminated: %+v\n", state.Terminated)
				c.l.Infof("container %s is terminated with %d status code at %v", pod.Name, state.Terminated.ExitCode, state.Terminated.FinishedAt)
				if app.State == model.ApplicationStateDeleting {
					if event.Type == watch.Deleted {
						if _, err := c.controller.ApplicationRepo.DeleteByID(ctx, app.ID); err != nil {
							c.l.Errorf("error deleting application %s: %v", app.ID.Hex(), err)
						}
					} else {
						c.l.Info("application is being deleted or already crashed, passing")
					}
					continue
				}
				c.l.Info("unexpected container death, notifying user")
				app.State = model.ApplicationStateCrashed
				if _, err := c.controller.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
					c.l.Errorf("error updating application: %v", err)
					continue
				}
			}

			if state.Waiting != nil {
				c.l.Debugf("state waiting: %+v\n", state.Waiting)
				c.l.Infof("container %s is waiting: %s", pod.Name, state.Waiting.Reason)
				switch state.Waiting.Reason {
				case "ContainerCreating":
					c.l.Infof("creating container %s", pod.Name)
					if app.Service == nil || app.Service.Deployment == nil {
						c.l.Warnf("application %s has no service or deployment, probably currently being created, ignoring", app.Name)
						continue
					}
					if app.Service.Deployment.CurrentPodName != "" {
						c.l.Warnf("application %s already has a running container, deleting it", app.Name)
						if err := c.controller.ServiceManager.DeletePod(ctx, pod.Namespace, app.Service.Deployment.CurrentPodName); err != nil {
							c.l.Errorf("error deleting pod: %v", err)
							continue
						}
					}
					app.Service.Deployment.CurrentPodName = pod.Name
					if _, err := c.controller.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
						c.l.Errorf("error updating application: %v", err)
						continue
					}

				case "CrashLoopBackOff":
					c.l.Infof("container %s is in crash loop, deleting to reset restart counter", pod.Name)
					c.l.Warnf("crashLoopBackOff, not restarting counter to not starve the cluster")
					// if err := c.controller.ServiceManager.DeletePod(ctx, pod.Namespace, pod.Name); err != nil {
					// 	c.l.Errorf("error deleting pod: %v", err)
					// 	continue
					// }
					// if app.Service == nil || app.Service.Deployment == nil {
					// 	c.l.Warnf("application %s has no service or deployment, probably currently being created, ignoring", app.Name)
					// 	continue
					// }
					// app.Service.Deployment.CurrentPodName = ""
					// if _, err := c.controller.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
					// 	c.l.Errorf("error updating application: %v", err)
					// 	continue
					// }

				default:
					c.l.Warnf("unknown way to handle waiting state %s for container %s", state.Waiting.Reason, pod.Name)
				}
			}
		}
	}
}
