package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type (
	Config struct {
		App      `yaml:"app"`
		Log      `yaml:"logger"`
		JWT      `yaml:"jwt"`
		Oauth    `yaml:"oauth"`
		RMQ      `yaml:"rabbitmq"`
		Database `yaml:"database"`
	}

	App struct {
		Name       string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version    string `env-required:"true" yaml:"version" env:"APP_VERSION"`
		Deployment string `env-required:"true" yaml:"deployment" env:"APP_DEPLOYMENT"`
	}

	Log struct {
		Level string `env-required:"true" yaml:"level" env:"LOG_LEVEL"`
		Type  string `env-required:"true" yaml:"type"  env:"LOG_TYPE"`
	}

	JWT struct {
		Secret string `env:"JWT_SECRET"`
	}

	Oauth struct {
		Provider     string `env-required:"true" yaml:"provider"     env:"OAUTH_PROVIDER"`
		RedirectUri  string `env-required:"true" yaml:"redirectUri" env:"OAUTH_REDIRECT_URI"`
		CallbackUri  string `env-required:"true" yaml:"callbackUri" env:"OAUTH_CALLBACK_URI"`
		ClientId     string `env:"OAUTH_CLIENT_ID"`
		ClientSecret string `env:"OAUTH_CLIENT_SECRET"`
	}

	RMQ struct {
		URI           string `env-required:"true" yaml:"uri" env:"RABBITMQ_URI"`
		RequestQueue  string `env-required:"true" yaml:"requestQueue" env:"RABBITMQ_REQUEST_QUEUE"`
		ResponseQueue string `env-required:"true" yaml:"responseQueue" env:"RABBITMQ_REPONSE_QUEUE"`
	}

	Database struct {
		Driver string `env-required:"true"  yaml:"driver" env:"DATABASE_DRIVER"`
		URI    string `env:"DATABASE_URI"`
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{}

	if err := godotenv.Load("./config/.env"); err != nil {
		if err.Error() != "open ./config/.env: no such file or directory" {
			return nil, err
		} else {
			logrus.Warn(".env file not found, using env variables")
		}
	}

	if err := cleanenv.ReadConfig("./config/config.yml", cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}

	if cfg.Database.Driver != "mock" {
		if cfg.Database.URI == "" {
			return nil, fmt.Errorf("DATABASE_URI is not set, this env variable is required when using a non mock driver")
		}
	}

	mustCheck := []string{"JWT_SECRET", "OAUTH_CLIENT_ID", "OAUTH_CLIENT_SECRET"}

	for _, v := range mustCheck {
		if os.Getenv(v) == "" {
			return nil, fmt.Errorf("%s is not set, this env variable is required", v)
		}
	}

	return cfg, nil
}
