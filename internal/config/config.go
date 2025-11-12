package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Firebase    Firebase
	S3          S3
	Prometheus  Prometheus
	Probe       Probe
	Password    Password
	JWT         JWT
	Postgres    Postgres
	Redis       Redis
	HTTP        HTTP
	Log         Log
	Asynq       Asynq
	Application Application
	Debug       bool `env:"DEBUG" envDefault:"false"`
}

func Load() (Config, error) {
	var config Config

	if err := env.Parse(&config); err != nil {
		return Config{}, fmt.Errorf("env.Parse: %w", err)
	}

	// https://www.dannyguo.com/blog/how-to-use-newlines-in-an-environment-variable-file-for-docker
	config.JWT.PrivateKey = correctNewlines(config.JWT.PrivateKey)
	config.JWT.PublicKey = correctNewlines(config.JWT.PublicKey)
	config.Firebase.PrivateKey = correctNewlines(config.Firebase.PrivateKey)

	return config, nil
}

func correctNewlines(s string) string {
	return strings.NewReplacer(`"`, "", `\n`, "\n").Replace(s)
}
