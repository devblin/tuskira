package config

import "os"

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string

	QueueProvider     string
	SchedulerProvider string

	JWTSecret string
	JWTExpiry string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "tuskira"),
		ServerPort: getEnv("SERVER_PORT", "8080"),

		QueueProvider:     getEnv("QUEUE_PROVIDER", "river"),
		SchedulerProvider: getEnv("SCHEDULER_PROVIDER", "river"),

		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiry: getEnv("JWT_EXPIRY", "24h"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
