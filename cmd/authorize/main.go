package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/goccy/go-json"
	"github.com/labstack/gommon/random"
	"golang.org/x/oauth2"

	"github.com/SlashNephy/annict2anilist/config"
	"github.com/SlashNephy/annict2anilist/external/anilist"
	"github.com/SlashNephy/annict2anilist/external/annict"
	"github.com/SlashNephy/annict2anilist/logger"
)

func main() {
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", slog.Any("err", err))
		panic(err)
	}
	logger.SetLevel(cfg.LogLevel)

	if err = authorize(ctx, anilist.NewOAuth2Config(cfg), filepath.Join(cfg.TokenDirectory, "token-anilist.json")); err != nil {
		slog.Error("failed to authorize AniList client", slog.Any("err", err))
		panic(err)
	}
	slog.Info("authorized AniList client")

	if err = authorize(ctx, annict.NewOAuth2Config(cfg), filepath.Join(cfg.TokenDirectory, "token-annict.json")); err != nil {
		slog.Error("failed to authorize Annict client", slog.Any("err", err))
		panic(err)
	}
	slog.Info("authorized Annict client")
}

func authorize(ctx context.Context, config *oauth2.Config, path string) error {
	state := random.String(64)
	url := config.AuthCodeURL(state)
	slog.Info("open URL in browser, then paste code", slog.String("url", url))

	// stdin -> Code
	var code string
	// Windows does not allow long characters
	// workaround: https://github.com/golang/go/issues/42551
	in := bufio.NewReader(os.Stdin)
	if _, err := fmt.Fscan(in, &code); err != nil {
		return errors.WithStack(err)
	}

	// Code -> Token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return errors.WithStack(err)
	}

	// Token -> JSON
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return errors.WithStack(err)
	}

	// Save Token JSON
	if err = os.WriteFile(path, tokenJson, 0666); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
