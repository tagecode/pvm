package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/ui"
)

var aliasCmd = &cobra.Command{
	Use:   "alias [name] [version]",
	Short: "Manage version aliases",
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}

		switch len(args) {
		case 0:
			items, err := a.Aliases.List()
			if err != nil {
				handleErr(cmd, err)
				return
			}
			if ui.JSONEnabled(cmd) {
				_ = ui.PrintJSON(map[string]any{"aliases": items})
				return
			}
			if len(items) == 0 {
				ui.PrintLine("No aliases defined.")
				return
			}
			for name, target := range items {
				ui.PrintLine("%s -> %s", name, target)
			}
		case 2:
			if err := a.Aliases.Set(args[0], args[1]); err != nil {
				handleErr(cmd, err)
				return
			}
			if !ui.QuietEnabled(cmd) {
				id, _ := a.Aliases.Get(args[0])
				ui.PrintLine("Alias %s -> %s", args[0], id)
			}
		default:
			handleErr(cmd, app.UserError("usage: pvm alias [name] [version]", ""))
		}
	},
}

var unaliasCmd = &cobra.Command{
	Use:   "unalias <name>",
	Short: "Remove a version alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		if err := a.Aliases.Delete(args[0]); err != nil {
			handleErr(cmd, err)
			return
		}
		if !ui.QuietEnabled(cmd) {
			ui.PrintLine("Removed alias %s", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(aliasCmd)
	rootCmd.AddCommand(unaliasCmd)
}
