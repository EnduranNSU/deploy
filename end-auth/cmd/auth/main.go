package main

import (
	"auth/internal/app"
	"auth/internal/logging"

	_ "github.com/joho/godotenv/autoload"
	"github.com/num30/config"
	"github.com/rs/zerolog/log"
)

func init() {
	logging.SetupLogger(logging.Config{
		Level: "info",
		Console: logging.ConsoleLoggerConfig{
			Enable:   true,
			Encoding: "text",
		},
		File: logging.FileLoggerConfig{
			Enable: false,
		},
	})
}

// @title           Enduran Auth API
// @version         1.0
// @description     Сервис аутентификации Enduran (регистрация, логин, refresh, сброс пароля)
// @BasePath        /

// @schemes         http

// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	var cfg app.Config
	cfgName := app.GetConfigName()
	if err := config.NewConfReader(cfgName).WithPrefix("APP").Read(&cfg); err != nil {
		log.Fatal().Stack().Err(err).Msg("failed to load config")
	}

	logging.SetupLogger(cfg.Logger)

	srv, err := app.BuildServer(cfg)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("failed to build server")
	}

	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("http server stopped")
	}
}
