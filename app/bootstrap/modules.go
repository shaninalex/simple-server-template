package bootstrap

import (
	"log/slog"
	"simple-server-template/app/core"
	"simple-server-template/app/modules/blog"
	"simple-server-template/app/platform/clock"
	"simple-server-template/app/platform/httpclient"
)

type buildModulesDeps struct {
	Log      *slog.Logger
	Clock    clock.Clock
	PostsAPI *httpclient.Client
}

func buildModules(d buildModulesDeps) ([]core.Module, error) {
	blogModule := blog.New(blog.Deps{
		Log:      d.Log,
		Clock:    d.Clock,
		PostsAPI: d.PostsAPI,
	})

	modules := []core.Module{
		blogModule,
	}

	return modules, nil
}
