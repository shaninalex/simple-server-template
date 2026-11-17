package core

import "simple-server-template/app/core/bus"

// App is main transport abstraction. Everything happened through the App.
// Everything else just an environment, tools, adapters... (database, api, cache, email sender ...)
type App struct {
	Queries  *bus.QueryBus
	Commands *bus.CommandBus
	Events   *bus.EventBus
}

type Module interface {
	Name() string
	Register(a *App) error
}

// Resolve - lazyload App. Since app initialized inside PersistentPreRunE when we call it
// app will be nil when we register other commands. So we have to lazy load app.
// Why? When we call "--help" command or read other docs we do not need to start any infrastructure or background tasks.
type Resolve func() *App
