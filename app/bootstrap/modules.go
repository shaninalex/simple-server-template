package bootstrap

import (
	"log/slog"
	"simple-server-template/app/core"
	"simple-server-template/app/modules/blog"
	"simple-server-template/app/platform/clock"
)

type buildModulesDeps struct {
	Log   *slog.Logger
	Clock clock.Clock
}

func buildModules(d buildModulesDeps) ([]core.Module, error) {
	blogModule := blog.New(blog.Deps{
		Log:   d.Log,
		Clock: d.Clock,
	})

	modules := []core.Module{
		blogModule,
	}

	return modules, nil
}
