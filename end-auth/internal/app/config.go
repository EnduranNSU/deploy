package app

import (
	"strings"

	"auth/internal/logging"
	"auth/internal/service"

	"auth/internal/utils/env"

	"github.com/spf13/viper"
)

func GetConfigName() string {
	configPath := env.GetEnvWithDefault("APP_CONFIG_FILE", "config/config.yaml")
	oldnew := make([]string, 2*len(viper.SupportedExts))
	for i, ext := range viper.SupportedExts {
		oldnew[2*i] = "." + ext
		oldnew[2*i+1] = ""
	}
	return strings.NewReplacer(oldnew...).Replace(configPath)
}

type Config struct {
	HTTP   HTTPConfig     `mapstructure:"http"`
	DB     DBConfig       `mapstructure:"db"`
	Logger logging.Config `mapstructure:"logger"`
	Svc    service.Config `mapstructure:"svc"`
}

type HTTPConfig struct {
	Addr string `mapstructure:"addr" default:":8081"`
}

type DBConfig struct {
	DSN string `mapstructure:"dsn" default:"postgres://postgres:postgres@localhost:5432/auth?sslmode=disable"`
}
