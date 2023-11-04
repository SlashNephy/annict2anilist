package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v10"
	"github.com/cockroachdb/errors"
	"github.com/joho/godotenv"
)

type Config struct {
	AnnictClientID      string `env:"ANNICT_CLIENT_ID,required"`
	AnnictClientSecret  string `env:"ANNICT_CLIENT_SECRET,required"`
	AniListClientID     string `env:"ANILIST_CLIENT_ID,required"`
	AniListClientSecret string `env:"ANILIST_CLIENT_SECRET,required"`
	TokenDirectory      string `env:"TOKEN_DIRECTORY" envDefault:"."`
	DryRun              bool   `env:"DRY_RUN"`
	LogLevel            string `env:"LOG_LEVEL"`
}

func LoadConfig() (*Config, error) {
	envFile := flag.String("env-file", ".env", "path to .env file")
	flag.Parse()

	// .env がある場合だけ読み込む
	if _, err := os.Stat(*envFile); !os.IsNotExist(err) {
		if err = godotenv.Load(*envFile); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, errors.WithStack(err)
	}

	return &cfg, nil
}
