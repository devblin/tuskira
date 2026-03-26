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
	EventKey          string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "tuskira"),
		ServerPort: getEnv("SERVER_PORT", "8080"),

		QueueProvider:     getEnv("QUEUE_PROVIDER", "inngest"),
		SchedulerProvider: getEnv("SCHEDULER_PROVIDER", "inngest"),
		EventKey:          getEnv("EVENT_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
