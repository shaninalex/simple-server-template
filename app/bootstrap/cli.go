package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const instanceShutdownTimeout = 10 * time.Second

func NewCli() (*cobra.Command, func() *Instance) {
	var app *Instance
	instance := func() *Instance { return app }

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

	root.AddCommand(
		serveCmd(instance),
	)

	return root, instance
}

func RunCLI() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root, instance := NewCli()
	runErr := root.ExecuteContext(ctx)

	stop()

	code := 0
	if runErr != nil {
		fmt.Fprintln(os.Stderr, runErr)
		code = 1
	}
	if app := instance(); app != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), instanceShutdownTimeout)
		defer cancel()
		if err := app.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintln(os.Stderr, "shutdown:", err)
			code = 1
		}
	}
	return code
}
