package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL          string
	RedisURL             string
	RedisPassword        string
	JWTSecret            string
	JWTTTLSeconds        int
	SessionTTLSeconds    int
	ArrivalToleranceMeters float64
	MaxPlanResults       int
	BusesActiveDefaultLimit int
	BusesActiveMaxLimit  int
	LiveETAMaxStreamMinutes int
	Port                 string
}

// Load reads environment variables into a typed Config struct.
// it panics at startup if JWT_SECRET, DATABASE_URL, or REDIS_URL
// are empty since there is no safe fallback for these values.
func Load() *Config {
	// silently skip if no .env file present, picks up system env vars otherwise
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:            envOr("DATABASE_URL", "postgres://transit_app:transit_dev_password@localhost:5432/transit?sslmode=disable"),
		RedisURL:               envOr("REDIS_URL", "localhost:6379"),
		RedisPassword:          os.Getenv("REDIS_PASSWORD"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		JWTTTLSeconds:          envIntOr("JWT_TTL_SECONDS", 3600),
		SessionTTLSeconds:      envIntOr("SESSION_TTL_SECONDS", 60),
		ArrivalToleranceMeters: envFloatOr("ARRIVAL_TOLERANCE_METERS", 50),
		MaxPlanResults:         envIntOr("MAX_PLAN_RESULTS", 3),
		BusesActiveDefaultLimit: envIntOr("BUSES_ACTIVE_DEFAULT_LIMIT", 50),
		BusesActiveMaxLimit:    envIntOr("BUSES_ACTIVE_MAX_LIMIT", 200),
		LiveETAMaxStreamMinutes: envIntOr("LIVE_ETA_MAX_STREAM_MINUTES", 45),
		Port:                   envOr("PORT", "8080"),
	}

	// fail fast rather than serving traffic with a missing secret or no db
	if cfg.JWTSecret == "" {
		panic("JWT_SECRET is required but was not set")
	}
	if cfg.DatabaseURL == "" {
		panic("DATABASE_URL is required but was not set")
	}
	if cfg.RedisURL == "" {
		panic("REDIS_URL is required but was not set")
	}

	return cfg
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envFloatOr(key string, fallback float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func (c *Config) RedisAddr() string {
	return c.RedisURL
}
