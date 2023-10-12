package anilist

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
		ClientID:     config.AniListClientID,
		ClientSecret: config.AniListClientSecret,
		RedirectURL:  "https://anilist.co/api/v2/oauth/pin",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://anilist.co/api/v2/oauth/authorize",
			TokenURL: "https://anilist.co/api/v2/oauth/token",
		},
	}

	client, err := external.NewOAuth2Client(ctx, oauth, config, tokenFile)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: graphql.NewClient("https://graphql.anilist.co", client),
	}, nil
}
