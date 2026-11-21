package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"simple-server-template/app/core"
	"simple-server-template/app/core/bus"
	"simple-server-template/app/platform/clock"
	"simple-server-template/app/platform/httpclient"
	"simple-server-template/app/platform/telemetry"
)

// Instance - is the application instance. It links everything together
type Instance struct {
	Config    Config
	Core      *core.App
	Log       *slog.Logger
	Clock     clock.Clock
	Telemetry *telemetry.Telemetry

	modules []core.Module
}

func New(cfg Config) (inst *Instance, err error) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	clk := clock.System()

	tel, err := telemetry.New(context.Background(), cfg.Telemetry)
	if err != nil {
		return nil, fmt.Errorf("telemetry: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tel.Shutdown(context.Background()))
		}
	}()

	mp := tel.MeterProvider()
	commandMetrics, err := telemetry.BusMetrics(mp, telemetry.BusCommand)
	if err != nil {
		return nil, err
	}
	queryMetrics, err := telemetry.BusMetrics(mp, telemetry.BusQuery)
	if err != nil {
		return nil, err
	}
	eventMetrics, err := telemetry.BusMetrics(mp, telemetry.BusEvent)
	if err != nil {
		return nil, err
	}

	app := &core.App{
		Commands: bus.NewCommandBus(commandMetrics),
		Queries:  bus.NewQueryBus(queryMetrics),
		Events:   bus.NewEventBus(eventMetrics),
	}

	postsAPI, err := httpclient.New(httpclient.Config{
		BaseURL:   cfg.PostsAPI.BaseURL,
		Timeout:   cfg.PostsAPI.Timeout,
		Retries:   cfg.PostsAPI.Retries,
		UserAgent: "simple-server-template",
	})
	if err != nil {
		return nil, fmt.Errorf("posts api client: %w", err)
	}

	modules, err := buildModules(buildModulesDeps{log, clk, postsAPI})
	if err != nil {
		panic(err)
	}

	for _, m := range modules {
		if err := m.Register(app); err != nil {
			return nil, fmt.Errorf("register module %s: %w", m.Name(), err)
		}
	}

	app.Commands.Seal()
	app.Queries.Seal()
	app.Events.Seal()

	return &Instance{
		Config:    cfg,
		Core:      app,
		Log:       log,
		Clock:     clk,
		Telemetry: tel,

		modules: modules,
	}, nil
}

func (i *Instance) Shutdown(ctx context.Context) error {
	return i.Telemetry.Shutdown(ctx)
}
