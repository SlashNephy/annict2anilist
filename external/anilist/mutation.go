package anilist

import (
	"context"

	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"

	"github.com/SlashNephy/annict2anilist/domain/status"
)

type SaveMediaListEntryMutation struct {
	SaveMediaListEntry struct {
		ID int `graphql:"id"`
	} `graphql:"SaveMediaListEntry(mediaId: $mediaID, status: $status, progress: $progress)"`
}

type MediaListEntryUpdate struct {
	MediaID  int
	Status   status.AniListMediaListStatus
	Progress int
}

func (c *Client) SaveMediaListEntry(ctx context.Context, update *MediaListEntryUpdate) error {
	var mutation SaveMediaListEntryMutation
	variables := map[string]any{
		"mediaID":  update.MediaID,
		"status":   MediaListStatus(update.Status),
		"progress": update.Progress,
	}
	if err := c.client.Mutate(ctx, &mutation, variables); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *Client) BatchSaveMediaListEntry(ctx context.Context, updates []*MediaListEntryUpdate) error {
	eg, egctx := errgroup.WithContext(ctx)

	for _, u := range updates {
		u := u
		eg.Go(func() error {
			if err := c.SaveMediaListEntry(egctx, u); err != nil {
				return errors.WithStack(err)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

type DeleteMediaListEntryMutation struct {
	DeleteMediaListEntry struct {
		Deleted bool `graphql:"deleted"`
	} `graphql:"DeleteMediaListEntry(id: id)"`
}

func (c *Client) DeleteMediaListEntry(ctx context.Context, id int) error {
	var mutation DeleteMediaListEntryMutation
	variables := map[string]any{
		"id": id,
	}
	if err := c.client.Mutate(ctx, &mutation, variables); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
