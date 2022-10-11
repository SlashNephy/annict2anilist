package main

import (
	"context"
	"sync"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

type AniListClient struct {
	*graphql.Client
}

func NewAniListClient(ctx context.Context, config *Config, tokenFile string) (*AniListClient, error) {
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

	return &AniListClient{
		graphql.NewClient("https://graphql.anilist.co", client),
	}, nil
}

type AniListViewerQuery struct {
	Viewer struct {
		ID   int    `graphql:"id"`
		Name string `graphql:"name"`
	} `graphql:"Viewer"`
}

func (client *AniListClient) FetchViewer(ctx context.Context) (*AniListViewerQuery, error) {
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
	Status   MediaListStatus `graphql:"status"`
	Progress int             `graphql:"progress"`
	Media    struct {
		ID    int `graphql:"id"`
		IDMal int `graphql:"idMal"`
		Title struct {
			Native string `graphql:"native"`
		} `graphql:"title"`
	} `graphql:"media"`
}

type MediaListSort string

func (client *AniListClient) FetchLibrary(ctx context.Context, userID, chunk int, status MediaListStatus) (*AniListLibraryQuery, error) {
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
	} `graphql:"SaveMediaListEntry(mediaId: $mediaId, status: $status, progress: $progress)"`
}

func (client *AniListClient) CreateMediaStatus(ctx context.Context, mediaID int, status MediaListStatus, progress int) error {
	var q AniListCreateMediaStatusQuery
	v := map[string]any{
		"mediaId":  mediaID,
		"status":   status,
		"progress": progress,
	}
	if err := client.Mutate(ctx, &q, v); err != nil {
		return err
	}

	return nil
}

type AniListUpdateMediaStatusQuery struct {
	UpdateMediaListEntries []struct {
		ID int `graphql:"id"`
	} `graphql:"UpdateMediaListEntries(ids: [$entryID], status: $status, progress: $progress)"`
}

func (client *AniListClient) UpdateMediaStatus(ctx context.Context, entryID int, status MediaListStatus, progress int) error {
	var q AniListUpdateMediaStatusQuery
	v := map[string]any{
		"entryID":  entryID,
		"status":   status,
		"progress": progress,
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

func (client *AniListClient) DeleteMediaStatus(ctx context.Context, entryID int) error {
	var q AniListDeleteMediaStatusQuery
	v := map[string]any{
		"entryId": entryID,
	}
	if err := client.Mutate(ctx, &q, v); err != nil {
		return err
	}

	return nil
}

func (client *AniListClient) FetchAllEntries(ctx context.Context, userID int) ([]AniListLibraryEntry, error) {
	var eg errgroup.Group
	var mutex sync.Mutex
	var entries []AniListLibraryEntry

	for _, s := range []MediaListStatus{AniListCurrentStatus, AniListCompletedStatus, AniListPlanningStatus, AniListPausedStatus, AniListDroppedStatus, AniListRepeatingStatus} {
		status := s
		eg.Go(func() error {
			var chunk = 0
			for {
				library, err := client.FetchLibrary(ctx, userID, chunk, status)
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
