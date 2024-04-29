package controller

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

func (c *Controller) GetLogs(ctx context.Context, namespace string, app string, from string, to string) (*model.LogBlock, error) {
	c.l.Infof("getting logs for app=%s in namespace=%s from=%s to=%s", app, namespace, from, to)
	logBlock, err := c.logProvider.GetLogs(ctx, namespace, app, from, to)
	if err != nil {
		c.l.Errorf("failed to get logs for app=%s in namespace=%s from=%s to=%s: %v", app, namespace, from, to, err)
		return nil, err
	}
	return logBlock, nil
}
