package main

import (
	"github.com/caarlos0/env/v10"
)

type config struct {
	Endpoint string `env:"ENDPOINT"`
}

func loadConfig() (config, error) {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return config{}, err
	}
	return cfg, nil
}
