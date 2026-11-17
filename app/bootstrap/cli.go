package bootstrap

import (
	"simple-server-template/app/core"

	"github.com/spf13/cobra"
)

func NewCli() *cobra.Command {
	var app *Instance

	root := &cobra.Command{
		Use:           "server",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			var err error
			app, err = New()
			return err
		},
	}

	resolve := core.Resolve(func() *core.App { return app.Core })
	root.AddCommand(
		serveCmd(resolve),
	)

	return root
}
