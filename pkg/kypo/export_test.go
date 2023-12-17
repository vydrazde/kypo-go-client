package kypo

import "context"

var (
	RefreshToken = func(c *Client, ctx context.Context) error { return c.refreshToken(ctx) }
	Authenticate = func(c *Client) error { return c.authenticate() }
)
