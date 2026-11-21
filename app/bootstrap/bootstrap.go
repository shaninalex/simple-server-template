package bootstrap

import (
	"fmt"
	"log/slog"
	"os"
	"simple-server-template/app/core"
	"simple-server-template/app/core/bus"
	"simple-server-template/app/platform/clock"
	"simple-server-template/app/platform/httpclient"
)

// Instance - is the application instance. It links everything together
type Instance struct {
	Core  *core.App
	Log   *slog.Logger
	Clock clock.Clock

	modules []core.Module
}

func New(cfg Config) (*Instance, error) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	clk := clock.System()

	// ======= Core =======
	var write []bus.Middleware // TODO: add transaction middleware. Docs: https://go.dev/doc/database/execute-transactions
	var read []bus.Middleware
	app := &core.App{
		Commands: bus.NewCommandBus(write...),
		Queries:  bus.NewQueryBus(read...),
		Events:   bus.NewEventBus(),
	}

	// ======= Remote data sources =======
	postsAPI, err := httpclient.New(httpclient.Config{
		BaseURL:   cfg.PostsAPI.BaseURL,
		Timeout:   cfg.PostsAPI.Timeout,
		Retries:   cfg.PostsAPI.Retries,
		UserAgent: "simple-server-template",
	})
	if err != nil {
		return nil, fmt.Errorf("posts api client: %w", err)
	}

	// modules
	modules, err := buildModules(buildModulesDeps{log, clk, postsAPI})
	if err != nil {
		panic(err)
	}

	for _, m := range modules {
		if err := m.Register(app); err != nil {
			return nil, fmt.Errorf("register module %s: %w", m.Name(), err)
		}
	}

	// stop registering handlers and listeners
	app.Commands.Seal()
	app.Queries.Seal()
	app.Events.Seal()

	return &Instance{
		Core:  app,
		Log:   log,
		Clock: clk,

		modules: modules,
	}, nil
}
