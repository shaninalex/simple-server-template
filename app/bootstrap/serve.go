package bootstrap

import (
	"log"
	"net/http"
	"simple-server-template/app/adapters/api"
	"simple-server-template/app/adapters/httpx"
	"simple-server-template/app/adapters/httpx/middlewares"
	"simple-server-template/app/core"

	"github.com/spf13/cobra"
)

func serveCmd(resolve core.Resolve) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run http server",
		RunE: func(cmd *cobra.Command, args []string) error {
			router := httpx.NewServeMux()
			router.Use(middlewares.Logging)
			router.Use(middlewares.Recovery)

			api.RegisterAPIControllers(router, resolve)

			log.Println("Server started at :8080")

			return http.ListenAndServe(":8080", router)
			// TODO: graceful shutdown
		},
	}

	return cmd
}
