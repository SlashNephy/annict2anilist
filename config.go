package main

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
)

const UserAgent = "annict2anilist/1.0 (+https://github.com/SlashNephy/annict2anilist)"

type Config struct {
	AnnictClientID       string `env:"ANNICT_CLIENT_ID,required"`
	AnnictClientSecret   string `env:"ANNICT_CLIENT_SECRET,required"`
	AniListClientID      string `env:"ANILIST_CLIENT_ID,required"`
	AniListClientSecret  string `env:"ANILIST_CLIENT_SECRET,required"`
	TokenDirectory       string `env:"TOKEN_DIRECTORY" envDefault:"."`
	IntervalMinutes      int64  `env:"INTERVAL_MINUTES"`
	DryRun               bool   `env:"DRY_RUN"`
	IgnoredAnnictIds     []int  `env:"-"`
	RawIgnoredAnnictIds  string `env:"IGNORED_ANNICT_IDS"`
	IgnoredAniListIds    []int  `env:"-"`
	RawIgnoredAniListIds string `env:"IGNORED_ANILIST_IDS"`
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

	if cfg.RawIgnoredAnnictIds != "" {
		for _, rawId := range strings.Split(cfg.RawIgnoredAnnictIds, ",") {
			id, err := strconv.Atoi(rawId)
			if err != nil {
				return nil, err
			}

			cfg.IgnoredAnnictIds = append(cfg.IgnoredAnnictIds, id)
		}
	}

	if cfg.RawIgnoredAniListIds != "" {
		for _, rawId := range strings.Split(cfg.RawIgnoredAniListIds, ",") {
			id, err := strconv.Atoi(rawId)
			if err != nil {
				return nil, err
			}

			cfg.IgnoredAniListIds = append(cfg.IgnoredAniListIds, id)
		}
	}

	return cfg, nil
}
