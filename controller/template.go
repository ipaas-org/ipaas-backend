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
	app.IsPublic = false
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

	c.l.Debugf("envs: %v", app.Envs)

	user, err := c.UserRepo.FindByCode(ctx, userCode)
	if err != nil {
		c.l.Errorf("error finding user by code: %v", err)
		return nil, err
	}

	labels := c.GenerateLabels(name, userCode, app.Kind)

	switch template.Kind {
	case model.ApplicationKindStorage:
		if !c.IsNameAvailableUserWide(ctx, name, userCode) {
			return nil, ErrApplicationNameNotAvailable
		}
		app.DnsName = name
	case model.ApplicationKindManagment:
		if !c.IsNameAvailableSystemWide(ctx, name) {
			return nil, ErrApplicationNameNotAvailable
		}
		//todo: could also be a random string but for now lets just use the name
		host := fmt.Sprintf("%s.%s", app.Name, c.app.BaseDefaultDomain)
		app.DnsName = host

		labels = append(labels, c.GenerateTraefikDnsLables(name, host, app.ListeningPort)...)
	default:
		return nil, ErrUnsupportedApplicationKind
	}

	if err := c.InsertApplication(ctx, app); err != nil {
		c.l.Errorf("error inserting application: %v", err)
		return nil, err
	}
	c.l.Debugf("template: %+v", template)
	container, err := c.createConnectAndStartContainer(ctx, name, template.ImageID, user.NetworkID, app.Envs, labels)
	if err != nil {
		c.l.Errorf("error creating container: %v", err)
		app.State = model.ApplicationStateFailed
		if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
			c.l.Errorf("error updating application: %v", err)
		}
		return nil, err
	}

	app.State = model.ApplicationStateRunning
	app.Container = container
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return nil, err
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
