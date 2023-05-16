package controller

import (
	"github.com/ipaas-org/ipaas-backend/config"
	"github.com/ipaas-org/ipaas-backend/pkg/jwt"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/ipaas-org/ipaas-backend/services/oauth"
	"github.com/ipaas-org/ipaas-backend/services/oauth/github"
	"github.com/ipaas-org/ipaas-backend/services/serviceManager"
	"github.com/ipaas-org/ipaas-backend/services/serviceManager/docker"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	l *logrus.Logger

	userRepo        repo.UserRepoer
	tokenRepo       repo.TokenRepoer
	stateRepo       repo.StateRepoer
	applicationRepo repo.ApplicationRepoer

	oauthService   oauth.Oauther
	jwtHandler     *jwt.JWThandler
	serviceManager serviceManager.ServiceManager
	app            config.App

	rabbitUri        string
	requestQueueName string
}

func NewBuilderController(config *config.Config, l *logrus.Logger) *Controller {
	var provider oauth.Oauther
	switch config.Oauth.Provider {
	case oauth.ProviderGithub:
		oauth := config.Oauth
		provider = github.NewGithubOauth(oauth.ClientId, oauth.ClientSecret, oauth.CallbackUri)
	default:
		l.Fatalf("Unknown oauth provider: %s", config.Oauth.Provider)
	}

	serviceManager, err := docker.NewDockerApplicationManager()
	if err != nil {
		l.Fatalf("Failed to create docker service manager: %v", err)
	}

	jwtHandler := jwt.NewJWThandler(config.JWT.Secret, config.App.Name+":"+config.App.Version)

	return &Controller{
		l:                l,
		oauthService:     provider,
		jwtHandler:       jwtHandler,
		serviceManager:   serviceManager,
		app:              config.App,
		rabbitUri:        config.RMQ.URI,
		requestQueueName: config.RMQ.RequestQueue,
	}
}

func (c *Controller) SetUserRepo(userRepo repo.UserRepoer) {
	c.userRepo = userRepo
}

func (c *Controller) SetTokenRepo(tokenRepo repo.TokenRepoer) {
	c.tokenRepo = tokenRepo
}

func (c *Controller) SetStateRepo(stateRepo repo.StateRepoer) {
	c.stateRepo = stateRepo
}

func (c *Controller) SetApplicationRepo(applicationRepo repo.ApplicationRepoer) {
	c.applicationRepo = applicationRepo
}
