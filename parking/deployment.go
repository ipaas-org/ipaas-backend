package parking

// package controller

// import (
// 	"context"

// 	"github.com/ipaas-org/ipaas-backend/model"
// )

// func (c *Controller) CreateNewDeployment(ctx context.Context, app *model.Application) error {
// 	var err error
// 	switch app.Kind {
// 	case model.ApplicationKindWeb:
// 		err = c.createWebDeployment(ctx, app)
// 	case model.ApplicationKindDatabase:
// 		err = c.createDbDeployment(ctx, app)
// 	default:
// 		c.l.Errorf("unknown application kind: %s", app.Kind)
// 		err = ErrUnsupportedApplicationKind
// 	}
// 	return err
// }

// func (c *Controller) createWebDeployment(ctx context.Context, app *model.Application) error {
// 	//insert into database
// 	err := c.InsertApplication(ctx, app)
// 	if err != nil {
// 		return err
// 	}

// 	return ErrNotImplemented
// }

// func (c *Controller) createDbDeployment(ctx context.Context, app *model.Application) error {
// 	return ErrNotImplemented
// }
