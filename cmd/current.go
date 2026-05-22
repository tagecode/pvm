package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/ui"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show active PHP version",
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		cur, err := a.Switch.Current()
		if err != nil {
			handleErr(cmd, err)
			return
		}
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(toCurrentJSON(a.Home, cur))
			return
		}
		ui.PrintLine("%s (%s)", cur.Version.String(), cur.Path)
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
