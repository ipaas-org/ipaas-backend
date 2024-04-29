package logprovider

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

type LogProvider interface {
	GetLogs(ctx context.Context, namespace string, app string, from string, to string) (*model.LogBlock, error)
}

const (
	LogProviderGrafanaLoki = "grafana-loki"
	LogProviderMock        = "mock"
)
