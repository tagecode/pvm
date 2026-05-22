package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/internal/config"
	"tagecode/pvm/pkg/ui"
)

var linkModeCmd = &cobra.Command{
	Use:   "link-mode [junction|symlink|copy]",
	Short: "Show or set the PHP version switch link mode",
	Long: `Control how pvm switches the active PHP version.

  junction  Directory junction (default, no admin required)
  symlink   Directory symlink (requires Developer Mode or admin)
  copy      Full directory copy (slow, most compatible)`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := config.Load(); err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}

		if len(args) == 0 {
			mode := config.LinkMode()
			if ui.JSONEnabled(cmd) {
				_ = ui.PrintJSON(map[string]string{"link_mode": mode})
				return
			}
			ui.PrintLine("link mode: %s", mode)
			return
		}

		mode := args[0]
		if dryRun(cmd) {
			if ui.JSONEnabled(cmd) {
				_ = ui.PrintJSON(map[string]string{"link_mode": mode, "dry_run": "true"})
				return
			}
			ui.PrintLine("Would set link mode to %s", mode)
			return
		}

		if err := config.SetLinkMode(mode); err != nil {
			handleErr(cmd, app.UserError(err.Error(), "use: pvm link-mode junction"))
			return
		}

		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(map[string]string{"link_mode": config.LinkMode()})
			return
		}
		if !ui.QuietEnabled(cmd) {
			ui.PrintLine("Link mode set to %s", config.LinkMode())
		}
	},
}

func init() {
	rootCmd.AddCommand(linkModeCmd)
}
