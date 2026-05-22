package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/internal/config"
	"tagecode/pvm/pkg/ui"
)

var cliVersion = "0.1.0"

type runtime struct {
	Home string
}

func initRuntime() (*runtime, error) {
	home, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &runtime{Home: home}, nil
}

func initApp() (*App, error) {
	home, err := config.Load()
	if err != nil {
		return nil, err
	}
	return NewApp(home), nil
}

func handleErr(cmd *cobra.Command, err error) {
	var exitErr *app.ExitError
	if errors.As(err, &exitErr) {
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(map[string]any{
				"error": exitErr.Message,
				"hint":  exitErr.Hint,
			})
		} else {
			ui.PrintError(errors.New(exitErr.Message))
			ui.PrintHint(exitErr.Hint)
		}
		os.Exit(exitErr.Code)
	}

	if ui.JSONEnabled(cmd) {
		_ = ui.PrintJSON(map[string]string{"error": err.Error()})
	} else {
		ui.PrintError(err)
	}
	os.Exit(app.ExitSystem)
}

func dryRun(cmd *cobra.Command) bool {
	v, _ := cmd.Root().PersistentFlags().GetBool("dry-run")
	return v
}
