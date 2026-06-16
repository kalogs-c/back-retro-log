package config

import (
	"os"
)

type Config struct {
	Addr     string
	DBPath   string
	RAWGKey  string
}

func Load() Config {
	cfg := Config{
		Addr:    ":3000",
		DBPath:  "data/backlog.db",
		RAWGKey: os.Getenv("RAWG_API_KEY"),
	}
	if addr := os.Getenv("ADDR"); addr != "" {
		cfg.Addr = addr
	}
	if path := os.Getenv("DB_PATH"); path != "" {
		cfg.DBPath = path
	}
	return cfg
}
