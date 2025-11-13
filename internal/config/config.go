package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Postgres Postgres
	HTTP     HTTP
	Debug    bool `env:"DEBUG" envDefault:"false"`
}

func Load() (Config, error) {
	var config Config

	if err := env.Parse(&config); err != nil {
		return Config{}, fmt.Errorf("env.Parse: %w", err)
	}
	return config, nil
}

func correctNewlines(s string) string {
	return strings.NewReplacer(`"`, "", `\n`, "\n").Replace(s)
}
