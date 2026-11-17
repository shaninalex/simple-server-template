package blog

import (
	"errors"
	"log/slog"
	"simple-server-template/app/core"
	"simple-server-template/app/core/bus"
	"simple-server-template/app/modules/blog/internal/handlers"
	"simple-server-template/app/platform/clock"
)

type Deps struct {
	Log   *slog.Logger
	Clock clock.Clock
}

type Module struct {
	deps Deps

	// queries
	postList *handlers.PostList
}

func New(d Deps) *Module {
	return &Module{
		deps: d,

		// queries
		postList: handlers.NewPostList(),
	}
}

func (m *Module) Name() string { return "blog" }

func (m *Module) Register(a *core.App) error {
	return errors.Join(
		bus.RegisterQuery(a.Queries, m.postList.Handle),
	)
}
