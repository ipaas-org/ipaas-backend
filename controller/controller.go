package controller

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/config"
	"github.com/ipaas-org/ipaas-backend/pkg/jwt"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/ipaas-org/ipaas-backend/services/imageBuilder"
	"github.com/ipaas-org/ipaas-backend/services/imageBuilder/ipaas"
	"github.com/ipaas-org/ipaas-backend/services/oauth"
	"github.com/ipaas-org/ipaas-backend/services/oauth/github"
	"github.com/ipaas-org/ipaas-backend/services/serviceManager"
	"github.com/ipaas-org/ipaas-backend/services/serviceManager/docker"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	l *logrus.Logger

	UserRepo        repo.UserRepoer
	TokenRepo       repo.TokenRepoer
	StateRepo       repo.StateRepoer
	ApplicationRepo repo.ApplicationRepoer

	oauthService   oauth.Oauther
	jwtHandler     *jwt.JWThandler
	serviceManager serviceManager.ServiceManager
	app            config.App
	imageBuilder   imageBuilder.ImageBuilder
}

func NewController(ctx context.Context, config *config.Config, l *logrus.Logger) *Controller {
	var provider oauth.Oauther
	switch config.Oauth.Provider {
	case oauth.ProviderGithub:
		oauth := config.Oauth
		provider = github.NewGithubOauth(oauth.ClientId, oauth.ClientSecret, oauth.CallbackUri)
	default:
		l.Fatalf("Unknown oauth provider: %s", config.Oauth.Provider)
	}

	serviceManager, err := docker.NewDockerApplicationManager(ctx)
	if err != nil {
		l.Fatalf("Failed to create docker service manager: %v", err)
	}

	imageBuilder := ipaas.NewIpaasImageBuilder(config.RMQ.URI, config.RMQ.RequestQueue)

	if config.JWT.Duration == 0 {
		config.JWT.Duration = jwt.DefaultExpirationTime
	}
	l.Info("jwt expiration is:", config.JWT.Duration)
	jwtHandler := jwt.NewJWThandler(config.JWT.Secret, config.App.Name+":"+config.App.Version, config.JWT.Duration)

	l.Infof("sending request on %s queue", config.RMQ.RequestQueue)
	return &Controller{
		l:              l,
		oauthService:   provider,
		jwtHandler:     jwtHandler,
		serviceManager: serviceManager,
		app:            config.App,
		imageBuilder:   imageBuilder,
	}
}
