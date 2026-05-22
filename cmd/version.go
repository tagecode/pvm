package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/pkg/ui"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print PVM version",
	Run: func(cmd *cobra.Command, args []string) {
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(map[string]string{"version": cliVersion})
			return
		}
		ui.PrintLine("pvm %s", cliVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
