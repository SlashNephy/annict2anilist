package annict

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"

	"github.com/SlashNephy/annict2anilist/domain/status"
)

type LibraryQuery struct {
	Viewer struct {
		LibraryEntries struct {
			Nodes []struct {
				Work Work `graphql:"work"`
			} `graphql:"nodes"`
			PageInfo struct {
				HasNextPage bool   `graphql:"hasNextPage"`
				EndCursor   string `graphql:"endCursor"`
			} `graphql:"pageInfo"`
		} `graphql:"libraryEntries(states: $states, after: $after)"`
	} `graphql:"viewer"`
}

type Work struct {
	ID                string                   `graphql:"id"`
	AnnictID          int                      `graphql:"annictId"`
	MALAnimeID        string                   `graphql:"malAnimeId"`
	SyobocalTID       int                      `graphql:"syobocalTid"`
	Title             string                   `graphql:"title"`
	NoEpisodes        bool                     `graphql:"noEpisodes"`
	ViewerStatusState status.AnnictStatusState `graphql:"viewerStatusState"`
	Episodes          struct {
		Nodes []Episode `graphql:"nodes"`
	} `graphql:"episodes"`
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
	eg, egctx := errgroup.WithContext(ctx)
	var mutex sync.Mutex
	var works []Work

	// MEMO: Annict は複数の states を渡せるが、まとめて渡すよりも1つずつ送った方が速い
	for _, s := range []status.AnnictStatusState{status.AnnictWatching, status.AnnictWatched, status.AnnictWannaWatch, status.AnnictOnHold, status.AnnictStopWatching} {
		s := s
		eg.Go(func() error {
			var after string
			for {
				library, err := c.FetchLibrary(egctx, []status.AnnictStatusState{s}, after)
				if err != nil {
					return errors.WithStack(err)
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
		return nil, errors.WithStack(err)
	}

	return works, nil
}
