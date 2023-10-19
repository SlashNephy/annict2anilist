package annict

import (
	"context"

	"github.com/cockroachdb/errors"
)

type ViewerQuery struct {
	Viewer User `graphql:"viewer"`
}

type User struct {
	Name     string `graphql:"name"`
	Username string `graphql:"username"`
}

func (c *Client) FetchViewer(ctx context.Context) (*ViewerQuery, error) {
	var query ViewerQuery
	if err := c.client.Query(ctx, &query, nil); err != nil {
		return nil, errors.WithStack(err)
	}

	return &query, nil
}
