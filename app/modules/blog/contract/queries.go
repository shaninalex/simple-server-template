package contract

import (
	"context"
	"simple-server-template/app/core"
	"simple-server-template/app/core/bus"
	"uuid"
)

type PostList struct {
	Limit, Offset int
}

func (PostList) Permission() (action string, scope int) { return "", 0 }

func AskPostList(ctx context.Context, a *core.App, q PostList) ([]PostView, error) {
	return bus.Ask[PostList, []PostView](ctx, a.Queries, q)
}

type PostDetail struct {
	ID uuid.UUID
}

func (PostDetail) Permission() (action string, scope int) { return "", 0 }

func AskPostDetail(ctx context.Context, a *core.App, q PostDetail) (PostView, error) {
	return bus.Ask[PostDetail, PostView](ctx, a.Queries, q)
}
