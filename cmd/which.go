package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/ui"
)

var whichCmd = &cobra.Command{
	Use:   "which <version>",
	Short: "Print installation path for a version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		p, err := a.Switch.Which(args[0])
		if err != nil {
			handleErr(cmd, err)
			return
		}
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(map[string]string{"path": p})
			return
		}
		ui.PrintLine("%s", p)
	},
}

func init() {
	rootCmd.AddCommand(whichCmd)
}
