package blog

import (
	"errors"
	"log/slog"
	"simple-server-template/app/core"
	"simple-server-template/app/core/bus"
	"simple-server-template/app/modules/blog/internal/datasource/postapi"
	"simple-server-template/app/modules/blog/internal/handlers"
	"simple-server-template/app/platform/clock"
	"simple-server-template/app/platform/httpclient"
)

type Deps struct {
	Log      *slog.Logger
	Clock    clock.Clock
	PostsAPI *httpclient.Client
}

type Module struct {
	deps Deps

	// queries
	postList   *handlers.PostList
	postDetail *handlers.PostDetail

	// commands
	postCreate *handlers.CreatePost
}

func New(d Deps) *Module {
	posts := postapi.New(d.PostsAPI)

	return &Module{
		deps: d,

		// queries
		postList:   handlers.NewPostList(posts),
		postDetail: handlers.NewPostDetail(posts),

		// commands
		postCreate: handlers.NewCreatePost(posts, d.Clock),
	}
}

func (m *Module) Name() string { return "blog" }

func (m *Module) Register(a *core.App) error {
	return errors.Join(
		bus.RegisterQuery(a.Queries, m.postList.Handle),
		bus.RegisterQuery(a.Queries, m.postDetail.Handle),
		bus.RegisterCommand(a.Commands, m.postCreate.Handle),
	)
}
