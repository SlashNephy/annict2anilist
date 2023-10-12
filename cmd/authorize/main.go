package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
	"github.com/labstack/gommon/random"
	"golang.org/x/oauth2"

	"github.com/SlashNephy/annict2anilist/config"
	"github.com/SlashNephy/annict2anilist/external/anilist"
	"github.com/SlashNephy/annict2anilist/external/annict"
)

func main() {
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	if err = authorize(ctx, anilist.NewOAuth2Config(cfg), filepath.Join(cfg.TokenDirectory, "token-anilist.json")); err != nil {
		panic(err)
	}
	log.Printf("Authorized AniList client")

	if err = authorize(ctx, annict.NewOAuth2Config(cfg), filepath.Join(cfg.TokenDirectory, "token-annict.json")); err != nil {
		panic(err)
	}
	log.Printf("Authorized Annict client")
}

func authorize(ctx context.Context, config *oauth2.Config, path string) error {
	state := random.String(64)
	url := config.AuthCodeURL(state)
	log.Printf("Open URL in browser, then paste code:\n=> %s\n", url)

	// stdin -> Code
	var code string
	// Windows does not allow long characters
	// workaround: https://github.com/golang/go/issues/42551
	in := bufio.NewReader(os.Stdin)
	if _, err := fmt.Fscan(in, &code); err != nil {
		panic(err)
	}

	// Code -> Token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return err
	}

	// Token -> JSON
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return err
	}

	// Save Token JSON
	return os.WriteFile(path, tokenJson, 0666)
}
