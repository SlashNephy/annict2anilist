package annict

import (
	"context"
	"log/slog"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/SlashNephy/annict2anilist/domain/status"
)

type LibraryQuery struct {
	Viewer struct {
		LibraryEntries LibraryEntryConnection `graphql:"libraryEntries(after: $after, first: $first, states: $states)"`
	} `graphql:"viewer"`
}

type LibraryEntryConnection struct {
	Edges    []LibraryEntryEdge `graphql:"edges"`
	PageInfo PageInfo           `graphql:"pageInfo"`
}

type LibraryEntryEdge struct {
	Node LibraryEntry `graphql:"node"`
}

type PageInfo struct {
	HasNextPage bool   `graphql:"hasNextPage"`
	EndCursor   string `graphql:"endCursor"`
}

type LibraryEntry struct {
	Work Work `graphql:"work"`
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

func (c *Client) FetchLibrary(ctx context.Context, states []status.AnnictStatusState, after string, first int) (*LibraryQuery, error) {
	var statuses []StatusState
	for _, s := range states {
		statuses = append(statuses, StatusState(s))
	}

	var query LibraryQuery
	variables := map[string]any{
		"after":  after,
		"first":  first,
		"states": statuses,
	}
	if err := c.client.Query(ctx, &query, variables); err != nil {
		return nil, errors.WithStack(err)
	}

	return &query, nil
}

func (c *Client) FetchAllWorks(ctx context.Context) ([]Work, error) {
	var (
		works []Work
		after string
	)
	for {
		library, err := c.FetchLibrary(ctx, []status.AnnictStatusState{
			status.AnnictWatching,
			status.AnnictWatched,
			status.AnnictWannaWatch,
			status.AnnictOnHold,
			status.AnnictStopWatching,
		}, after, 100)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, edge := range library.Viewer.LibraryEntries.Edges {
			works = append(works, edge.Node.Work)
		}
		slog.Info("fetch works", slog.Int("total", len(works)))

		if !library.Viewer.LibraryEntries.PageInfo.HasNextPage {
			return works, nil
		}

		after = library.Viewer.LibraryEntries.PageInfo.EndCursor
		time.Sleep(5 * time.Second)
	}
}
