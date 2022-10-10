package main

import (
	"context"
	"sync"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

func NewAniListClient(ctx context.Context, config *Config, tokenFile string) (*graphql.Client, error) {
	oauth := &oauth2.Config{
		ClientID:     config.AniListClientID,
		ClientSecret: config.AniListClientSecret,
		RedirectURL:  "https://anilist.co/api/v2/oauth/pin",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://anilist.co/api/v2/oauth/authorize",
			TokenURL: "https://anilist.co/api/v2/oauth/token",
		},
	}

	client, err := NewOAuth2Client(ctx, oauth, config, tokenFile)
	if err != nil {
		return nil, err
	}

	return graphql.NewClient("https://graphql.anilist.co", client), nil
}

type AniListViewerQuery struct {
	Viewer struct {
		ID   int    `graphql:"id"`
		Name string `graphql:"name"`
	} `graphql:"Viewer"`
}

func FetchAniListViewer(ctx context.Context, client *graphql.Client) (*AniListViewerQuery, error) {
	var q AniListViewerQuery
	if err := client.Query(ctx, &q, nil); err != nil {
		return nil, err
	}

	return &q, nil
}

type AniListLibraryQuery struct {
	MediaListCollection struct {
		Lists []struct {
			Entries []AniListLibraryEntry `graphql:"entries"`
		} `graphql:"lists"`
		HasNextChunk bool `graphql:"hasNextChunk"`
	} `graphql:"MediaListCollection(userId: $userId, sort: $sort, perChunk: $perChunk, chunk: $chunk, type: ANIME, forceSingleCompletedList: true, status: $status)"`
}

type AniListLibraryEntry struct {
	ID       int             `graphql:"id"`
	Progress int             `graphql:"progress"`
	Status   MediaListStatus `graphql:"status"`
	Score    float32         `graphql:"score"`
	Media    struct {
		ID    int `graphql:"id"`
		Title struct {
			Romaji string `graphql:"romaji"`
			Native string `graphql:"native"`
		} `graphql:"title"`
	} `graphql:"media"`
}

type MediaListSort string
type MediaListStatus string

var AniListMediaListStatuses = []MediaListStatus{"COMPLETED", "CURRENT", "DROPPED", "PAUSED", "PLANNING", "REPEATING"}

func FetchAniListLibrary(ctx context.Context, client *graphql.Client, userID, chunk int, status MediaListStatus) (*AniListLibraryQuery, error) {
	var q AniListLibraryQuery
	v := map[string]any{
		"userId":   userID,
		"sort":     []MediaListSort{"STARTED_ON"},
		"perChunk": 500,
		"chunk":    chunk,
		"status":   status,
	}
	if err := client.Query(ctx, &q, v); err != nil {
		return nil, err
	}

	return &q, nil
}

type AniListCreateMediaStatusQuery struct {
	SaveMediaListEntry struct {
		ID int `graphql:"id"`
	} `graphql:"SaveMediaListEntry(mediaId: $mediaId, status: $status, progress: $progress, score: $score)"`
}

func CreateAniListMediaStatus(ctx context.Context, client *graphql.Client, mediaID int, status MediaListStatus, progress int, score float32) error {
	var q AniListCreateMediaStatusQuery
	v := map[string]any{
		"mediaId":  mediaID,
		"status":   status,
		"progress": progress,
		"score":    score,
	}
	if err := client.Mutate(ctx, &q, v); err != nil {
		return err
	}

	return nil
}

type AniListUpdateMediaStatusQuery struct {
	UpdateMediaListEntries []struct {
		ID int `graphql:"id"`
	} `graphql:"UpdateMediaListEntries(ids: [$entryID], status: $status, progress: $progress, score: $score)"`
}

func UpdateAniListMediaStatus(ctx context.Context, client *graphql.Client, entryID int, status MediaListStatus, progress int, score float32) error {
	var q AniListUpdateMediaStatusQuery
	v := map[string]any{
		"entryID":  entryID,
		"status":   status,
		"progress": progress,
		"score":    score,
	}
	if err := client.Mutate(ctx, &q, v); err != nil {
		return err
	}

	return nil
}

type AniListDeleteMediaStatusQuery struct {
	DeleteMediaListEntry struct {
		Deleted bool `graphql:"deleted"`
	} `graphql:"DeleteMediaListEntry(id: $entryId)"`
}

func DeleteAniListMediaStatus(ctx context.Context, client *graphql.Client, entryID int) error {
	var q AniListDeleteMediaStatusQuery
	v := map[string]any{
		"entryId": entryID,
	}
	if err := client.Mutate(ctx, &q, v); err != nil {
		return err
	}

	return nil
}

func FetchAllAniListEntries(ctx context.Context, client *graphql.Client, userID int) ([]AniListLibraryEntry, error) {
	var eg errgroup.Group
	var mutex sync.Mutex
	var entries []AniListLibraryEntry

	for _, s := range AniListMediaListStatuses {
		status := s
		eg.Go(func() error {
			var chunk = 0
			for {
				library, err := FetchAniListLibrary(ctx, client, userID, chunk, status)
				if err != nil {
					return err
				}

				mutex.Lock()
				for _, list := range library.MediaListCollection.Lists {
					entries = append(entries, list.Entries...)
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
		return nil, err
	}

	return entries, nil
}
