package bootstrap

import (
	"fmt"
	"simple-server-template/app/core"

	"github.com/spf13/cobra"
)

func NewCli() *cobra.Command {
	var app *Instance

	root := &cobra.Command{
		Use:           "blogging",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := LoadConfig()
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}
			app, err = New(cfg)
			return err
		},
	}

	resolve := core.Resolve(func() *core.App { return app.Core })
	root.AddCommand(
		serveCmd(resolve),
	)

	return root
}

func RunCLI() int {
	root := NewCli()
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}
