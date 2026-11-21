package bootstrap

import (
	"simple-server-template/app/adapters/api"
	"simple-server-template/app/adapters/httpx"
	"simple-server-template/app/adapters/httpx/middlewares"
	"simple-server-template/app/adapters/web"
	"simple-server-template/app/core"
	"time"

	"github.com/spf13/cobra"
)

const httpShutdownTimeout = 15 * time.Second

func serveCmd(instance func() *Instance) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run http server",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := instance()
			resolve := core.Resolve(func() *core.App { return app.Core })

			metrics, err := middlewares.Metrics(app.Telemetry.MeterProvider())
			if err != nil {
				return err
			}

			router := httpx.NewServeMux()
			router.Use(middlewares.Logging)
			router.Use(metrics)
			router.Use(middlewares.Recovery)
			api.RegisterAPIControllers(router, resolve)
			web.RegisterWebRoutes(router, resolve)

			main, err := httpx.Listen("main", app.Config.HTTP.Addr, router)
			if err != nil {
				return err
			}
			admin, err := httpx.Listen("admin", app.Config.HTTP.AdminAddr, app.Telemetry.AdminHandler())
			if err != nil {
				_ = main.Close()
				return err
			}

			return httpx.Serve(cmd.Context(), app.Log, httpShutdownTimeout, main, admin)
		},
	}
}
