package anilist

import (
	"context"
	"github.com/SlashNephy/annict2anilist/domain/status"
)

type SaveMediaListEntryMutation struct {
	SaveMediaListEntry struct {
		ID int `graphql:"id"`
	} `graphql:"SaveMediaListEntry(mediaId: $mediaID, status: $status, progress: $progress)"`
}

func (c *Client) SaveMediaListEntry(ctx context.Context, mediaID int, status status.AniListMediaListStatus, progress int) error {
	var mutation SaveMediaListEntryMutation
	variables := map[string]any{
		"mediaID":  mediaID,
		"status":   status,
		"progress": progress,
	}
	if err := c.client.Mutate(ctx, &mutation, variables); err != nil {
		return err
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
		return err
	}

	return nil
}
