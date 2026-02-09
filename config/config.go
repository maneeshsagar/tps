package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Kafka    KafkaConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port int
}

type LogConfig struct {
	Level string
}

type PostgresConfig struct {
	Host                   string
	Port                   int
	User                   string
	Password               string
	DBName                 string
	SSLMode                string
	TimeZone               string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeMinutes int
}

type KafkaConfig struct {
	Brokers []string
}

func (p PostgresConfig) ConnMaxLifetime() time.Duration {
	return time.Duration(p.ConnMaxLifetimeMinutes) * time.Minute
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Postgres: PostgresConfig{
			Host:                   getEnv("POSTGRES_HOST", "localhost"),
			Port:                   getEnvInt("POSTGRES_PORT", 5432),
			User:                   getEnv("POSTGRES_USER", "postgres"),
			Password:               getEnv("POSTGRES_PASSWORD", "postgres"),
			DBName:                 getEnv("POSTGRES_DB", "tps"),
			SSLMode:                getEnv("POSTGRES_SSLMODE", "disable"),
			TimeZone:               getEnv("POSTGRES_TIMEZONE", "UTC"),
			MaxOpenConns:           getEnvInt("POSTGRES_MAX_OPEN_CONNS", 25),
			MaxIdleConns:           getEnvInt("POSTGRES_MAX_IDLE_CONNS", 5),
			ConnMaxLifetimeMinutes: getEnvInt("POSTGRES_CONN_MAX_LIFETIME_MINUTES", 5),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		},
	}
	return cfg, nil
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt returns the integer value of an environment variable or a default value
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		log.Printf("Warning: invalid integer value for %s, using default %d", key, defaultVal)
	}
	return defaultVal
}
