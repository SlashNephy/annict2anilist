package annict

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"

	"github.com/SlashNephy/annict2anilist/config"
	"github.com/SlashNephy/annict2anilist/external"
)

type Client struct {
	client *graphql.Client
}

func NewClient(ctx context.Context, config *config.Config, tokenFile string) (*Client, error) {
	oauth := &oauth2.Config{
		ClientID:     config.AnnictClientID,
		ClientSecret: config.AnnictClientSecret,
		Scopes:       []string{"read"},
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.annict.com/oauth/authorize",
			TokenURL: "https://api.annict.com/oauth/token",
		},
	}

	client, err := external.NewOAuth2Client(ctx, oauth, config, tokenFile)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: graphql.NewClient("https://api.annict.com/graphql", client),
	}, nil
}
