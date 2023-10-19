package annict

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/SlashNephy/annict2anilist/domain/status"
)

type LibraryQuery struct {
	Viewer struct {
		LibraryEntries struct {
			Edges []struct {
				Node struct {
					Work Work `graphql:"work"`
				} `graphql:"node"`
			} `graphql:"edges"`
			PageInfo struct {
				HasNextPage bool   `graphql:"hasNextPage"`
				EndCursor   string `graphql:"endCursor"`
			} `graphql:"pageInfo"`
		} `graphql:"libraryEntries(states: $states, after: $after)"`
	} `graphql:"viewer"`
}

type Work struct {
	AnnictID          int                      `graphql:"annictId"`
	MALAnimeID        string                   `graphql:"malAnimeId"`
	SyobocalTID       int                      `graphql:"syobocalTid"`
	Title             string                   `graphql:"title"`
	ViewerStatusState status.AnnictStatusState `graphql:"viewerStatusState"`
	NoEpisodes        bool                     `graphql:"noEpisodes"`
	Episodes          EpisodeConnection        `graphql:"episodes"`
}

type EpisodeConnection struct {
	Edges []EpisodeEdge `graphql:"edges"`
}

type EpisodeEdge struct {
	Node Episode `graphql:"node"`
}

type Episode struct {
	ViewerDidTrack bool `graphql:"viewerDidTrack"`
}

type StatusState status.AnnictStatusState

func (c *Client) FetchLibrary(ctx context.Context, states []status.AnnictStatusState, after string) (*LibraryQuery, error) {
	var statuses []StatusState
	for _, s := range states {
		statuses = append(statuses, StatusState(s))
	}

	var query LibraryQuery
	variables := map[string]any{
		"states": statuses,
		"after":  after,
	}
	if err := c.client.Query(ctx, &query, variables); err != nil {
		return nil, errors.WithStack(err)
	}

	return &query, nil
}

func (c *Client) FetchAllWorks(ctx context.Context) ([]Work, error) {
	var works []Work
	var after string
	for {
		library, err := c.FetchLibrary(ctx, []status.AnnictStatusState{
			status.AnnictWatching,
			status.AnnictWatched,
			status.AnnictWannaWatch,
			status.AnnictOnHold,
			status.AnnictStopWatching,
		}, after)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, node := range library.Viewer.LibraryEntries.Edges {
			works = append(works, node.Node.Work)
		}

		if !library.Viewer.LibraryEntries.PageInfo.HasNextPage {
			break
		}

		after = library.Viewer.LibraryEntries.PageInfo.EndCursor
		time.Sleep(3 * time.Second)
	}

	return works, nil
}
