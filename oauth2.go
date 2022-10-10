package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
	"github.com/labstack/gommon/random"
	"golang.org/x/oauth2"
)

func NewOAuth2Client(ctx context.Context, oauth *oauth2.Config, config *Config, tokenFile string) (*http.Client, error) {
	path := filepath.Join(config.TokenDirectory, tokenFile)

	var token *oauth2.Token
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		state := random.String(64)
		url := oauth.AuthCodeURL(state)
		fmt.Printf("Open URL & Paste code:\n=> %s\n", url)

		// stdin -> Code
		var code string
		_, err := fmt.Scanln(&code)
		if err != nil {
			return nil, err
		}

		// Code -> Token
		token, err = oauth.Exchange(ctx, code)
		if err != nil {
			return nil, err
		}

		// Token -> JSON
		tokenJson, err := json.Marshal(token)
		if err != nil {
			return nil, err
		}

		// Save Token JSON
		if err = os.WriteFile(path, tokenJson, 0666); err != nil {
			return nil, err
		}
	} else {
		// Load Token JSON
		tokenJson, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		// JSON -> Token
		if err = json.Unmarshal(tokenJson, &token); err != nil {
			return nil, err
		}
	}

	source := oauth.TokenSource(ctx, token)
	return oauth2.NewClient(ctx, source), nil
}
