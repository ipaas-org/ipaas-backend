package controller

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/config"
	"github.com/ipaas-org/ipaas-backend/pkg/jwt"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/ipaas-org/ipaas-backend/services/gitProvider"
	"github.com/ipaas-org/ipaas-backend/services/gitProvider/github"
	"github.com/ipaas-org/ipaas-backend/services/imageBuilder"
	"github.com/ipaas-org/ipaas-backend/services/imageBuilder/ipaas"
	logprovider "github.com/ipaas-org/ipaas-backend/services/logProvider"
	"github.com/ipaas-org/ipaas-backend/services/logProvider/grafana"
	k8smanager "github.com/ipaas-org/ipaas-backend/services/serviceManager/k8s"
	"github.com/sirupsen/logrus"
)

const (
	staticTempEnvironment = "prod"
)

type Controller struct {
	l *logrus.Logger

	UserRepo        repo.UserRepoer
	TokenRepo       repo.TokenRepoer
	StateRepo       repo.StateRepoer
	ApplicationRepo repo.ApplicationRepoer
	TemplateRepo    repo.TemplateRepoer
	TempTokenRepo   repo.TemporaryTokenStorage

	gitProvider    gitProvider.Provider
	jwtHandler     *jwt.JWThandler
	ServiceManager *k8smanager.K8sOrchestratedServiceManager
	app            config.App
	traefik        config.Traefik
	config         *config.Config
	imageBuilder   imageBuilder.ImageBuilder
	logProvider    logprovider.LogProvider
}

func NewController(ctx context.Context, config *config.Config, l *logrus.Logger) *Controller {
	var provider gitProvider.Provider
	switch config.GitProvider.Provider {
	case gitProvider.ProviderGithub:
		oauth := config.GitProvider
		oauth.CallbackUri = config.App.ApiUrl + oauth.CallbackUri
		provider = github.NewGithubOauth(oauth.ClientId, oauth.ClientSecret, oauth.CallbackUri)
	default:
		l.Fatalf("Unknown oauth provider: %s", config.GitProvider.Provider)
	}

	serviceManager, err := k8smanager.NewK8sOrchestratedServiceManager(
		config.K8s.KubeConfigPath,
		config.K8s.CPUResource,
		config.K8s.MemoryResource)
	if err != nil {
		l.Fatalf("Failed to create k8s service manager: %v", err)
	}

	imageBuilder := ipaas.NewIpaasImageBuilder(config.RMQ.URI, config.RMQ.RequestQueue)

	if config.JWT.Duration == 0 {
		config.JWT.Duration = jwt.DefaultExpirationTime
	}
	l.Info("jwt expiration is:", config.JWT.Duration)
	jwtHandler := jwt.NewJWThandler(config.JWT.Secret, config.App.Name+":"+config.App.Version, config.JWT.Duration)

	var logProvider logprovider.LogProvider
	switch config.LogProvider.Provider {
	case logprovider.LogProviderGrafanaLoki:
		l.Infof("Using Grafana Loki as log provider")
		logProvider, err = grafana.NewGrafanaLogProvider(ctx, config.LogProvider.Token, config.LogProvider.BaseUrl)
		if err != nil {
			l.Fatalf("Failed to create grafana loki log provider: %v", err)
		}
	default:
		l.Fatalf("Unknown log provider: %s", config.LogProvider.Provider)
	}

	l.Infof("sending request on %s queue", config.RMQ.RequestQueue)
	return &Controller{
		l:              l,
		gitProvider:    provider,
		jwtHandler:     jwtHandler,
		ServiceManager: serviceManager,
		app:            config.App,
		imageBuilder:   imageBuilder,
		config:         config,
		traefik:        config.Traefik,
		logProvider:    logProvider,
	}
}
