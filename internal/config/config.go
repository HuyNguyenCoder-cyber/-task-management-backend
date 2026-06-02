package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort           string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	RedisHost         string
	RedisPort         string
	JWTSecret         string
	JWTExpiresInHours int
}

func Load() (*Config, error) {
	_ = loadDotEnv()

	cfg := &Config{
		AppPort:           getEnv("APP_PORT", "8080"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "3306"),
		DBUser:            getEnv("DB_USER", "app_user"),
		DBPassword:        getEnv("DB_PASSWORD", "app_password"),
		DBName:            getEnv("DB_NAME", "task_management"),
		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         getEnv("REDIS_PORT", "6379"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTExpiresInHours: getEnvAsInt("JWT_EXPIRES_IN_HOURS", 24),
	}

	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	return cfg, nil
}

func loadDotEnv() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	for {
		envPath := filepath.Join(currentDir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return godotenv.Load(envPath)
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return nil
		}

		currentDir = parentDir
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

func (c *Config) RedisAddr() string {
	return net.JoinHostPort(c.RedisHost, c.RedisPort)
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsedValue
}
