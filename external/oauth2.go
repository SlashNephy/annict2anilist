package external

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/goccy/go-json"
	"golang.org/x/oauth2"

	"github.com/SlashNephy/annict2anilist/config"
)

func NewOAuth2Client(ctx context.Context, oauth *oauth2.Config, config *config.Config, tokenFile string) (*http.Client, error) {
	path := filepath.Join(config.TokenDirectory, tokenFile)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("token file not found: you need to run cmd/authorize first")
	}

	// Load Token JSON
	tokenJson, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// JSON -> Token
	var token *oauth2.Token
	if err = json.Unmarshal(tokenJson, &token); err != nil {
		return nil, errors.WithStack(err)
	}

	source := oauth.TokenSource(ctx, token)
	if !token.Valid() {
		slog.Info("refreshing token",
			slog.String("token_type", token.TokenType),
			slog.String("token_file", tokenFile),
		)

		token, err = source.Token()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// Token -> JSON
		tokenJson, err = json.Marshal(token)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// Save Token JSON
		if err = os.WriteFile(path, tokenJson, 0666); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	slog.Info("active token",
		slog.String("expiry", token.Expiry.String()),
		slog.String("token_type", token.TokenType),
		slog.String("token_file", tokenFile),
	)

	client := oauth2.NewClient(ctx, source)
	client.Timeout = 15 * time.Second
	return client, nil
}
