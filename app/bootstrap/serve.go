package bootstrap

import (
	"net/http"
	"simple-server-template/app/core"

	"github.com/spf13/cobra"
)

func serveCmd(resolve core.Resolve) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run http server",
		RunE: func(cmd *cobra.Command, args []string) error {
			router := http.NewServeMux()
			// Then `resolve` will be provided to the route handlers, and then into the command/query callers
			// So the command will be called inside initialized bus.
			return http.ListenAndServe(":8080", router)
		},
	}

	return cmd
}
