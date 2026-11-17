package contract

import (
	"context"
	"simple-server-template/app/core"
	"simple-server-template/app/core/bus"
)

type PostList struct {
	Limit, Offset int
}

func (PostList) Permission() (action string, scope int) { return "", 0 }

func AskPostList(ctx context.Context, a *core.App, q PostList) ([]PostView, error) {
	return bus.Ask[PostList, []PostView](ctx, a.Queries, q)
}
