package contract

import (
	"context"
	"simple-server-template/app/core"
	"simple-server-template/app/core/bus"
)

type CreatePost struct {
	Post PostView
}

func (CreatePost) Permission() (action string, scope int) { return "", 0 }

func ExecCreatePost(ctx context.Context, a *core.App, cmd CreatePost) (PostView, error) {
	return bus.Execute[CreatePost, PostView](ctx, a.Commands, cmd)
}
