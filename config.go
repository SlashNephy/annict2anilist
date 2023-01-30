package main

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
)

const UserAgent = "annict2anilist/1.0 (+https://github.com/SlashNephy/annict2anilist)"

type Config struct {
	AnnictClientID      string `env:"ANNICT_CLIENT_ID,required"`
	AnnictClientSecret  string `env:"ANNICT_CLIENT_SECRET,required"`
	AniListClientID     string `env:"ANILIST_CLIENT_ID,required"`
	AniListClientSecret string `env:"ANILIST_CLIENT_SECRET,required"`
	TokenDirectory      string `env:"TOKEN_DIRECTORY" envDefault:"."`
	IntervalMinutes     int64  `env:"INTERVAL_MINUTES"`
	DryRun              bool   `env:"DRY_RUN"`
}

func LoadConfig() (*Config, error) {
	envFile := flag.String("env-file", ".env", "path to .env file")
	flag.Parse()

	// .env がある場合だけ読み込む
	if _, err := os.Stat(*envFile); !os.IsNotExist(err) {
		if err = godotenv.Load(*envFile); err != nil {
			return nil, err
		}
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
