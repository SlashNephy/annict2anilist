package anilist

import "context"

type ViewerQuery struct {
	Viewer struct {
		ID   int    `graphql:"id"`
		Name string `graphql:"name"`
	} `graphql:"Viewer"`
}

func (c *Client) FetchViewer(ctx context.Context) (*ViewerQuery, error) {
	var query ViewerQuery
	if err := c.client.Query(ctx, &query, nil); err != nil {
		return nil, err
	}

	return &query, nil
}
