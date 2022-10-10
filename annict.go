package main

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

func NewAnnictClient(ctx context.Context, config *Config, tokenFile string) (*graphql.Client, error) {
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

	client, err := NewOAuth2Client(ctx, oauth, config, tokenFile)
	if err != nil {
		return nil, err
	}

	return graphql.NewClient("https://api.annict.com/graphql", client), nil
}
