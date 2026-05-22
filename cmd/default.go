package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/ui"
)

var defaultCmd = &cobra.Command{
	Use:   "default [version]",
	Short: "Show or set the default PHP version",
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		if len(args) == 0 {
			id, err := a.Aliases.Get("default")
			if err != nil {
				handleErr(cmd, err)
				return
			}
			if ui.JSONEnabled(cmd) {
				_ = ui.PrintJSON(map[string]string{"default": id})
				return
			}
			ui.PrintLine("default -> %s", id)
			return
		}
		resolved, err := a.ResolveID(args[0])
		if err != nil {
			handleErr(cmd, err)
			return
		}
		if err := a.Aliases.Set("default", resolved); err != nil {
			handleErr(cmd, err)
			return
		}
		if !ui.QuietEnabled(cmd) {
			ui.PrintLine("Default version set to %s", resolved)
		}
	},
}

func init() {
	rootCmd.AddCommand(defaultCmd)
}
