package config

import (
	"cmp"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
}

func Load() *Config {
	return &Config{
		Port:        cmp.Or(os.Getenv("PORT"), "8080"),
		DatabaseURL: cmp.Or(os.Getenv("DATABASE_URL"), "postgres://postgres:postgres@localhost:5432/bookshelf?sslmode=disable"),
		JWTSecret:   cmp.Or(os.Getenv("JWT_SECRET"), "dev-secret-change-me"),
	}
}
