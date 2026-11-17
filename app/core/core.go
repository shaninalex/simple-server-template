package core

type App struct {
}

type Module interface {
	Name() string
	Register(a *App) error
}

type Resolve func() *App
