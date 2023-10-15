package external

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/goccy/go-json"
	"golang.org/x/oauth2"

	"github.com/SlashNephy/annict2anilist/config"
)

func NewOAuth2Client(ctx context.Context, oauth *oauth2.Config, config *config.Config, tokenFile string) (*http.Client, error) {
	path := filepath.Join(config.TokenDirectory, tokenFile)

	var token *oauth2.Token
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("token file not found: you need to run cmd/authorize first")
	}

	// Load Token JSON
	tokenJson, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// JSON -> Token
	if err = json.Unmarshal(tokenJson, &token); err != nil {
		return nil, errors.WithStack(err)
	}

	source := oauth.TokenSource(ctx, token)
	return oauth2.NewClient(ctx, source), nil
}
