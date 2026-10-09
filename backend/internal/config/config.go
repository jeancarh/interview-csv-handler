package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DBPath         string
	SeedCSV        string
	MaxUploadBytes int64
	MaxImportRows  int
}

func Load() Config {
	return Config{
		Port:           env("PORT", "8080"),
		DBPath:         env("DB_PATH", "data/shop.db"),
		SeedCSV:        env("SEED_CSV", ""),
		MaxUploadBytes: int64(envInt("MAX_UPLOAD_MB", 5)) << 20,
		MaxImportRows:  envInt("MAX_IMPORT_ROWS", 10_000),
	}
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	n, err := strconv.Atoi(os.Getenv(key))
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
