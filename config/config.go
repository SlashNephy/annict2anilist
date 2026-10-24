package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
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
	var envPath string
	if f := flag.Lookup("env-file"); f != nil {
		envPath = f.Value.String()
	} else {
		flag.StringVar(&envPath, "env-file", ".env", "path to .env file")
	}

	if !flag.Parsed() {
		flag.Parse()
	}

	if envPath == "" {
		envPath = ".env"
	}

	// .env がある場合だけ読み込む
	if _, err := os.Stat(envPath); !os.IsNotExist(err) {
		if err = godotenv.Load(envPath); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, errors.WithStack(err)
	}

	return &cfg, nil
}
