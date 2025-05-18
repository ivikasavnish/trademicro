package config

import (
	"os"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port         string
	IsMicro      bool
	WorkerHost   string
	WorkerUser   string
	WorkerSSHKey string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	SecretKey []byte
}

// Load returns application configuration loaded from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvWithDefault("PORT", "8000"),
			IsMicro:      os.Getenv("SERVER_ROLE") == "micro",
			WorkerHost:   getEnvWithDefault("WORKER_HOST", "instance-20250416-112838"),
			WorkerUser:   getEnvWithDefault("WORKER_USER", "root"),
			WorkerSSHKey: getEnvWithDefault("WORKER_SSH_KEY", "/opt/trademicro/.ssh/worker_key"),
		},
		Database: DatabaseConfig{
			URL: getEnvWithDefault("POSTGRES_URL", "postgres://postgres:password@localhost:5432/trademicro"),
		},
		Redis: RedisConfig{
			URL: getEnvWithDefault("REDIS_URL", "redis://localhost:6379/0"),
		},
		JWT: JWTConfig{
			SecretKey: []byte(getEnvWithDefault("SECRET_KEY", "default_secret_key")),
		},
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
