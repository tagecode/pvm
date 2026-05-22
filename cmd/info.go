package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/ui"
)

var infoCmd = &cobra.Command{
	Use:   "info <version>",
	Short: "Show detailed information about a version",
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
		v, err := a.Registry.Get(id)
		if err != nil {
			handleErr(cmd, err)
			return
		}
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(v)
			return
		}
		ui.PrintLine("Version: %s", v.Version.String())
		ui.PrintLine("ID:      %s", v.ID)
		ui.PrintLine("Path:    %s", v.Path)
		ui.PrintLine("Active:  %v", v.Active)
		ui.PrintLine("Default: %v", v.Default)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
