package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Service  string
	Env      string
	AppPort  int
	GrpcPort int

	DatabaseURL string

	DBMaxConns        int
	DBMinConns        int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	JWTSecret         string
	JWTAccessTokenTTL time.Duration
}

func loadDotEnv() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = "configs/.env.local"
	}

	_ = godotenv.Load(envFile)
}

func Load() Config {
	loadDotEnv()

	return Config{
		Service:  env("SERVICE", "go-backend-template"),
		Env:      env("ENV", "development"),
		AppPort:  envInt("APP_PORT", 8080),
		GrpcPort: envInt("GRPC_PORT", 9090),

		DatabaseURL: must("DATABASE_URL"),

		DBMaxConns:        envInt("DB_MAX_CONNS", 25),
		DBMinConns:        envInt("DB_MIN_CONNS", 5),
		DBConnMaxLifetime: envDurationMinutes("DB_CONN_MAX_LIFETIME_MINUTES", 5),
		DBConnMaxIdleTime: envDurationMinutes("DB_CONN_MAX_IDLE_TIME_MINUTES", 5),

		JWTSecret:         must("JWT_SECRET"),
		JWTAccessTokenTTL: envDurationMinutes("JWT_ACCESS_TOKEN_TTL_MINUTES", 60),
	}
}

func must(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("%s is required", key))
	}

	return v
}

func env(key string, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("%s must be int", key))
	}

	return i
}

func envDurationMinutes(key string, def int) time.Duration {
	return time.Duration(envInt(key, def)) * time.Minute
}
