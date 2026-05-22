package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/ui"
)

var useCmd = &cobra.Command{
	Use:   "use <version>",
	Short: "Switch active PHP version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		id, err := a.ResolveID(args[0])
		if err != nil {
			handleErr(cmd, err)
			return
		}
		result, err := a.Switch.Use(cmd.Context(), id, app.UseOptions{DryRun: dryRun(cmd)})
		if err != nil {
			handleErr(cmd, err)
			return
		}
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(map[string]any{
				"active":          result.ID,
				"setup_applied":   result.SetupApplied,
				"session_refreshed": !dryRun(cmd),
			})
			return
		}
		if dryRun(cmd) {
			ui.PrintLine("Would switch to %s", result.ID)
			return
		}
		ui.PrintLine("%s", a.Switch.UseMessage(result))
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
