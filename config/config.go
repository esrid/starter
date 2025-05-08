package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Env       string
	Port      string
	Debug     bool
	Database  *Database
	SecretKey string
}
type Database struct {
	HOST     string
	USER     string
	DATABASE string
	PASSWORD string
	PORT     string
}

var (
	cfg  *Config
	once sync.Once
)

func Load() *Config {
	once.Do(func() {
		_ = godotenv.Load()

		cfg = &Config{
			Env:       getEnv("APP_ENV", "development"),
			Port:      getEnv("PORT", "8080"),
			Debug:     getEnvAsBool("DEBUG", true),
			SecretKey: getEnv("SECRET_KEY", "changeme"),
			Database: &Database{
				HOST:     getEnv("DB_HOST", "localhost"),
				USER:     getEnv("DB_USER", "user"),
				DATABASE: getEnv("DB_NAME", "dbname"),
				PASSWORD: getEnv("DB_PASSWORD", "dbpassword"),
				PORT:     getEnv("DB_PORT", "5432"),
			},
		}
	})
	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := os.Getenv(name)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func (d *Database) String() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", d.USER, d.PASSWORD, d.HOST, d.PORT, d.DATABASE)
}
