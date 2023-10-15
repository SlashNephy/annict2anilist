package anilist

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"

	"github.com/SlashNephy/annict2anilist/domain/status"
)

type LibraryQuery struct {
	MediaListCollection struct {
		Lists []struct {
			Entries []LibraryEntry `graphql:"entries"`
		} `graphql:"lists"`
		HasNextChunk bool `graphql:"hasNextChunk"`
	} `graphql:"MediaListCollection(userId: $userId, sort: $sort, perChunk: $perChunk, chunk: $chunk, type: ANIME, forceSingleCompletedList: true, status: $status)"`
}

type LibraryEntry struct {
	ID       int                           `graphql:"id"`
	Status   status.AniListMediaListStatus `graphql:"status"`
	Progress int                           `graphql:"progress"`
	Media    struct {
		ID    int `graphql:"id"`
		IDMal int `graphql:"idMal"`
		Title struct {
			Native string `graphql:"native"`
		} `graphql:"title"`
		Episodes int         `graphql:"episodes"`
		Status   MediaStatus `graphql:"status"`
	} `graphql:"media"`
}

type MediaStatus string

const MediaStatusFinished MediaStatus = "FINISHED"

type MediaListSort string

const MediaListSortStartedOn MediaListSort = "STARTED_ON"

type MediaListStatus status.AniListMediaListStatus

func (c *Client) FetchLibrary(ctx context.Context, userID, chunk int, status status.AniListMediaListStatus) (*LibraryQuery, error) {
	var query LibraryQuery
	variables := map[string]any{
		"userId":   userID,
		"sort":     []MediaListSort{MediaListSortStartedOn},
		"perChunk": 500,
		"chunk":    chunk,
		"status":   MediaListStatus(status),
	}
	if err := c.client.Query(ctx, &query, variables); err != nil {
		return nil, errors.WithStack(err)
	}

	return &query, nil
}

func (c *Client) FetchAllEntries(ctx context.Context, userID int) ([]LibraryEntry, error) {
	eg, egctx := errgroup.WithContext(ctx)
	var mutex sync.Mutex
	var entries []LibraryEntry

	for _, s := range []status.AniListMediaListStatus{status.AniListCurrent, status.AniListCompleted, status.AniListPlanning, status.AniListPaused, status.AniListDropped, status.AniListRepeating} {
		s := s
		eg.Go(func() error {
			var chunk = 0
			for {
				library, err := c.FetchLibrary(egctx, userID, chunk, s)
				if err != nil {
					return errors.WithStack(err)
				}

				mutex.Lock()
				for _, list := range library.MediaListCollection.Lists {
					for _, e := range list.Entries {
						entries = append(entries, e)
					}
				}
				mutex.Unlock()

				if !library.MediaListCollection.HasNextChunk {
					return nil
				}

				chunk++
			}
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, errors.WithStack(err)
	}

	return entries, nil
}
