package config

import (
	"github.com/joho/godotenv"
	"github.com/maneeshsagar/tps/logger"
)

func LoadEnv(log logger.Logger) {
	if err := godotenv.Load(); err != nil {
		log.Debug(".env file not found (optional)", "err", err)
	}
}
