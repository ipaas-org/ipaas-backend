package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
)

// todo: use this function to check if the name is available for a database
func (c *Controller) IsNameAvailableUserWide(ctx context.Context, name, userCode string) bool {
	_, err := c.ApplicationRepo.FindByNameAndOwner(ctx, name, userCode)
	available := err == repo.ErrNotFound
	c.l.Debugf("is name[%s] available for %s: %t", name, userCode, available)
	return available
}

func (c *Controller) CreateNewApplicationBasedOnTemplate(ctx context.Context, userCode, name string, template *model.Template, envs []model.KeyValue) (*model.Application, error) {
	c.l.Debugf("creating a new application for %s based on template %s", userCode, template.Code)
	app := new(model.Application)
	app.Name = name
	app.Kind = template.Kind
	app.State = model.ApplicationStateStarting
	app.CreatedAt = time.Now()
	app.Owner = userCode
	app.IsUpdatable = false
	app.ListeningPort = template.ListeningPort
	app.Envs = envs
	app.BasedOn = template.Code

	c.l.Debugf("default envs: %v", template.DefaultEnvs)
	if template.DefaultEnvs != nil {
		app.Envs = append(app.Envs, template.DefaultEnvs...)
	}

	for _, te := range template.RequiredEnvs {
		var found bool
		for _, e := range envs {
			if e.Key == te.Key {
				found = true
				break
			}
		}
		if !found {
			return nil, ErrMissingRequiredEnvForTemplate
		}
	}

	user, err := c.UserRepo.FindByCode(ctx, userCode)
	if err != nil {
		c.l.Errorf("error finding user by code: %v", err)
		return nil, err
	}

	switch template.Kind {
	case model.ApplicationKindStorage:
		if err := c.createNewStorageKindService(ctx, template, app, user); err != nil {
			c.l.Errorf("error creating storage service: %v", err)
			return nil, err
		}

	case model.ApplicationKindManagment:
		if err := c.createNewManagmentKindService(ctx, template, app, user); err != nil {
			c.l.Errorf("error creating managment service: %v", err)
			return nil, err
		}
	default:
		return nil, ErrUnsupportedApplicationKind
	}

	return app, nil
}

func (c *Controller) GetTemplateByCode(ctx context.Context, code string) (*model.Template, error) {
	template, err := c.TemplateRepo.FindByCode(ctx, code)
	if err != nil {
		c.l.Errorf("error finding template by code %s: %v", code, err)
		return nil, err
	}
	return template, nil
}

func (c *Controller) ListTemplates(ctx context.Context) ([]*model.Template, error) {
	templates, err := c.TemplateRepo.FindAllAvailable(ctx)
	if err != nil {
		c.l.Errorf("error finding templates: %v", err)
		return nil, err
	}
	return templates, nil
}

func (c *Controller) createNewStorageKindService(ctx context.Context, template *model.Template, app *model.Application, user *model.User) error {
	if !c.IsNameAvailableUserWide(ctx, app.Name, user.Code) {
		return ErrApplicationNameNotAvailable
	}
	app.DnsName = app.Name
	app.Visiblity = model.ApplicationVisiblityPrivate

	if err := c.InsertApplication(ctx, app); err != nil {
		c.l.Errorf("error inserting application: %v", err)
		return err
	}

	configMap, err := c.createConfigMap(ctx, app, user, app.Envs)
	if err != nil {
		return err
	}

	GiSize := int64(1 * 1024 * 1024 * 1024) // 1Gi
	storageClass := "longhorn-test"
	pvc, err := c.CreatePersistantVolumeClaim(ctx, app, user, storageClass, GiSize)
	if err != nil {
		return err
	}
	volume := new(model.Volume)
	volume.Name = fmt.Sprintf("vol-%s", app.Name)
	volume.MountPath = template.PersistancePath
	volume.PersistantVolumeClaim = pvc

	deployment, err := c.createDeployment(ctx, app, user, template.ImageName, configMap.Name, volume)
	if err != nil {
		return err
	}
	deployment.ConfigMap = configMap

	//watch for deployment ready state, it's non blocking so we just check at the end
	done, errChan := c.serviceManager.WaitDeploymentReadyState(ctx, user.Namespace, deployment.Name)
	var errWhileWaiting error
loop:
	for {
		select {
		case err := <-errChan:
			c.l.Errorf("error while waiting for deployment: %v", err)
			errWhileWaiting = err
			break loop
		case _, ok := <-done:
			if !ok {
				//todo: choose what to do in this case
				c.l.Errorf("done chan is closed, either context was cancelled or something worst O-O")
				if ctx.Err() != nil {
					c.l.Errorf("context cancelled while a deployment was being created, no further changes will be done at the moment")
				} else {
					errWhileWaiting = fmt.Errorf("internal error")
				}
			} else {
				c.l.Debugf("deployment %s is available", deployment.Name)
			}
			break loop
		}
	}

	service, err := c.createService(ctx, app, user)
	if err != nil {
		return err
	}
	service.Deployment = deployment

	if errWhileWaiting != nil {
		//todo: handle waiting error, it's probably because it reached a timeout
		//in this case it probably means that we reached a cpu/mem cap and we should
		//expand the infra, it should really not happen
		c.l.Errorf("internal error reached, this should not happen, check infrastructure resources left")
		app.State = model.ApplicationStateFailed
	} else {
		app.State = model.ApplicationStateRunning
		app.Service = service
	}
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}

	return nil
}

func (c *Controller) createNewManagmentKindService(ctx context.Context, template *model.Template, app *model.Application, user *model.User) error {
	if !c.IsNameAvailableSystemWide(ctx, app.Name) {
		return ErrApplicationNameNotAvailable
	}
	//todo: could also be a random string but for now lets just use the name
	host := fmt.Sprintf("%s.%s", app.Name, c.app.BaseDefaultDomain)
	app.DnsName = host
	app.Visiblity = model.ApplicationVisiblityPublic

	if err := c.InsertApplication(ctx, app); err != nil {
		c.l.Errorf("error inserting application: %v", err)
		return err
	}

	configMap, err := c.createConfigMap(ctx, app, user, app.Envs)
	if err != nil {
		return err
	}

	deployment, err := c.createDeployment(ctx, app, user, template.ImageName, configMap.Name, nil)
	if err != nil {
		return err
	}
	deployment.ConfigMap = configMap

	//watch for deployment ready state, it's non blocking so we just check at the end
	done, errChan := c.serviceManager.WaitDeploymentReadyState(ctx, user.Namespace, deployment.Name)
	var errWhileWaiting error
loop:
	for {
		select {
		case err := <-errChan:
			c.l.Errorf("error while waiting for deployment: %v", err)
			errWhileWaiting = err
			break loop
		case _, ok := <-done:
			if !ok {
				//todo: choose what to do in this case
				c.l.Errorf("done chan is closed, either context was cancelled or something worst O-O")
				if ctx.Err() != nil {
					c.l.Errorf("context cancelled while a deployment was being created, no further changes will be done at the moment")
				} else {
					errWhileWaiting = fmt.Errorf("internal error")
				}
			} else {
				c.l.Debugf("deployment %s is available", deployment.Name)
			}
			break loop
		}
	}

	service, err := c.createService(ctx, app, user)
	if err != nil {
		return err
	}
	service.Deployment = deployment

	ingressRoute, err := c.createIngressRoute(ctx, app, user, app.DnsName, service.Name, service.Port)
	if err != nil {
		return err
	}
	service.IngressRoute = ingressRoute

	if errWhileWaiting != nil {
		//todo: handle waiting error, it's probably because it reached a timeout
		//in this case it probably means that we reached a cpu/mem cap and we should
		//expand the infra, it should really not happen
		c.l.Errorf("internal error reached, this should not happen, check infrastructure resources left")
		app.State = model.ApplicationStateFailed
	} else {
		app.State = model.ApplicationStateRunning
		app.Service = service
	}
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}
	return nil
}
