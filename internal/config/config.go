package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"strings"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Postgres Postgres
	HTTP     HTTP
	Debug    bool `env:"DEBUG" envDefault:"false"`
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {

		log.Println("config.Load: no .env.dev file found, reading from environment variables")
	}
	var config Config

	if err := env.Parse(&config); err != nil {
		return Config{}, fmt.Errorf("env.Parse: %w", err)
	}
	return config, nil
}

func correctNewlines(s string) string {
	return strings.NewReplacer(`"`, "", `\n`, "\n").Replace(s)
}
