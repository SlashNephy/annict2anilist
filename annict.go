package main

import (
	"context"
	"sync"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

type AnnictClient struct {
	*graphql.Client
}

func NewAnnictClient(ctx context.Context, config *Config, tokenFile string) (*AnnictClient, error) {
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

	return &AnnictClient{
		graphql.NewClient("https://api.annict.com/graphql", client),
	}, nil
}

type AnnictViewerQuery struct {
	Viewer struct {
		Name     string `graphql:"name"`
		Username string `graphql:"username"`
	} `graphql:"viewer"`
}

func (client *AnnictClient) FetchViewer(ctx context.Context) (*AnnictViewerQuery, error) {
	var q AnnictViewerQuery
	if err := client.Query(ctx, &q, nil); err != nil {
		return nil, err
	}

	return &q, nil
}

type AnnictLibraryQuery struct {
	Viewer struct {
		LibraryEntries struct {
			Nodes []struct {
				Work AnnictWork `graphql:"work"`
			} `graphql:"nodes"`
			PageInfo struct {
				HasNextPage bool   `graphql:"hasNextPage"`
				EndCursor   string `graphql:"endCursor"`
			} `graphql:"pageInfo"`
		} `graphql:"libraryEntries(states: $states, after: $after)"`
	} `graphql:"viewer"`
}

type AnnictWork struct {
	ID                string      `graphql:"id"`
	AnnictID          int         `graphql:"annictId"`
	MALAnimeID        string      `graphql:"malAnimeId"`
	SyobocalTID       int         `graphql:"syobocalTid"`
	Title             string      `graphql:"title"`
	NoEpisodes        bool        `graphql:"noEpisodes"`
	ViewerStatusState StatusState `graphql:"viewerStatusState"`
	Episodes          struct {
		Nodes []AnnictEpisode `graphql:"nodes"`
	} `graphql:"episodes"`
}

type AnnictEpisode struct {
	ViewerDidTrack bool `graphql:"viewerDidTrack"`
}

func (client *AnnictClient) FetchLibrary(ctx context.Context, states []StatusState, after string) (*AnnictLibraryQuery, error) {
	var q AnnictLibraryQuery
	v := map[string]any{
		"states": states,
		"after":  after,
	}
	if err := client.Query(ctx, &q, v); err != nil {
		return nil, err
	}

	return &q, nil
}

func (client *AnnictClient) FetchAllWorks(ctx context.Context) ([]AnnictWork, error) {
	var eg errgroup.Group
	var mutex sync.Mutex
	var works []AnnictWork

	// MEMO: Annict は複数の states を渡せるが、まとめて渡すよりも1つずつ送った方が速い
	for _, s := range []StatusState{AnnictWatchingStatus, AnnictWatchedStatus, AnnictWannaWatchStatus, AnnictOnHoldStatus, AnnictStopWatching} {
		status := s
		eg.Go(func() error {
			var after string
			for {
				library, err := client.FetchLibrary(ctx, []StatusState{status}, after)
				if err != nil {
					return err
				}

				mutex.Lock()
				for _, node := range library.Viewer.LibraryEntries.Nodes {
					works = append(works, node.Work)
				}
				mutex.Unlock()

				if !library.Viewer.LibraryEntries.PageInfo.HasNextPage {
					return nil
				}

				after = library.Viewer.LibraryEntries.PageInfo.EndCursor
			}
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return works, nil
}
