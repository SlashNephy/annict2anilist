package annict

import "context"

type ViewerQuery struct {
	Viewer struct {
		Name     string `graphql:"name"`
		Username string `graphql:"username"`
	} `graphql:"viewer"`
}

func (c *Client) FetchViewer(ctx context.Context) (*ViewerQuery, error) {
	var query ViewerQuery
	if err := c.client.Query(ctx, &query, nil); err != nil {
		return nil, err
	}

	return &query, nil
}
