package main

import (
	"github.com/caarlos0/env/v10"
)

type config struct {
	Addr string `env:"ADDR" envDefault:"8080"`
}

func loadConfig() (config, error) {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return config{}, err
	}
	return cfg, nil
}
