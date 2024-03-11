package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type (
	Config struct {
		App         `yaml:"app"`
		Log         `yaml:"logger"`
		JWT         `yaml:"jwt"`
		GitProvider `yaml:"gitProvider"`
		RMQ         `yaml:"rabbitmq"`
		HTTP        `yaml:"http"`
		Database    `yaml:"database"`
		Traefik     `yaml:"traefik"`
		K8s         `yaml:"k8s"`
	}

	App struct {
		Name              string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version           string `env-required:"true" yaml:"version" env:"APP_VERSION"`
		Deployment        string `env-required:"true" yaml:"deployment" env:"APP_DEPLOYMENT"`
		ApiUrl            string `env-required:"true" yaml:"apiUrl" env:"APP_API_URL"`
		FrontendUrl       string `env-required:"true" yaml:"frontendUrl" env:"APP_FRONTEND_URL"`
		BaseDefaultDomain string `env-required:"true" yaml:"baseDefaultDomain" env:"APP_BASE_DEFAULT_DOMAIN"`
	}

	Log struct {
		Level string `env-required:"true" yaml:"level" env:"LOG_LEVEL"`
		Type  string `env-required:"true" yaml:"type"  env:"LOG_TYPE"`
	}

	JWT struct {
		Secret   string        `env-required:"true" env:"JWT_SECRET"`
		Duration time.Duration `yaml:"duration" env:"JWT_DURATION"`
	}

	GitProvider struct {
		Provider     string `env-required:"true" yaml:"provider"    env:"GIT_PROVIDER"`
		RedirectUri  string `env-required:"true" yaml:"redirectUri" env:"GIT_PROVIDER_REDIRECT_URI"`
		CallbackUri  string `env-required:"true" yaml:"callbackUri" env:"GIT_PROVIDER_CALLBACK_URI"`
		ClientId     string `env:"GIT_PROVIDER_CLIENT_ID"`
		ClientSecret string `env:"GIT_PROVIDER_CLIENT_SECRET"`
	}

	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	RMQ struct {
		URI           string `env-required:"true" yaml:"uri" env:"RABBITMQ_URI"`
		RequestQueue  string `env-required:"true" yaml:"requestQueue" env:"RABBITMQ_REQUEST_QUEUE"`
		ResponseQueue string `env-required:"true" yaml:"responseQueue" env:"RABBITMQ_REPONSE_QUEUE"`
	}

	Database struct {
		Driver string `env-required:"true" yaml:"driver" env:"DATABASE_DRIVER"`
		URI    string `env:"DATABASE_URI"`
	}

	Traefik struct {
		ApiBaseUrl string `env-required:"true" yaml:"apiBaseUrl" env:"TRAEFIK_API_BASE_URL"`
		Username   string `env:"TRAEFIK_USERNAME"`
		Password   string `env:"TRAEFIK_PASSWORD"`
	}

	K8s struct {
		KubeConfigPath   string `env-required:"true" yaml:"kubeConfigPath" env:"K8S_KUBE_CONFIG_PATH"`
		CPUResource      string `env-required:"true" yaml:"cpuResource" env:"K8S_CPU_RESOURCE"`
		MemoryResource   string `env-required:"true" yaml:"memoryResource" env:"K8S_MEMORY_RESOURCE"`
		RegistryUrl      string `env-required:"true" yaml:"registryUrl" env:"K8S_REGISTRY_URL"`
		RegistryUsername string `env-required:"true" yaml:"registryUsername" env:"K8S_REGISTRY_USERNAME"`
		RegistryPassword string `env-required:"true" yaml:"registryPassword" env:"K8S_REGISTRY_PASSWORD"`
	}
)

func NewConfig(configPath ...string) (*Config, error) {
	cfg := new(Config)

	path := "./"
	if len(configPath) > 0 {
		path = configPath[0]
	}

	if err := godotenv.Load(path + ".env"); err != nil {
		if err.Error() != "open "+path+".env: no such file or directory" {
			return nil, err
		} else {
			logrus.Warn(".env file not found, using env variables")
		}
	}

	mustCheck := []string{"JWT_SECRET", "GIT_PROVIDER_CLIENT_ID", "GIT_PROVIDER_CLIENT_SECRET", "TRAEFIK_USERNAME", "TRAEFIK_PASSWORD"}

	for _, v := range mustCheck {
		logrus.Debug(os.Getenv(v))
		if os.Getenv(v) == "" {
			return nil, fmt.Errorf("%s is not set, this env variable is required", v)
		}
	}

	if err := cleanenv.ReadConfig(path+"config.yml", cfg); err != nil {
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

	return cfg, nil
}
