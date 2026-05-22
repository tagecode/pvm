package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/ui"
)

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show PHP-related environment status",
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		st, err := a.Setup.Status()
		if err != nil {
			handleErr(cmd, err)
			return
		}
		ps, c := a.Setup.RefreshScript()
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(map[string]any{
				"pvm_home":      st.PVMHome,
				"php_home":      st.PHPHome,
				"phprc":         st.PHPRC,
				"in_path":       st.InPath,
				"path_at_front": st.PathAtFront,
				"path_index":    st.PathIndex,
				"user_path":     st.Path,
				"refresh": map[string]string{
					"powershell": ps,
					"cmd":        c,
				},
			})
			return
		}
		ui.PrintLine("PVM_HOME:  %s", st.PVMHome)
		ui.PrintLine("PHP_HOME:  %s", st.PHPHome)
		ui.PrintLine("PHPRC:     %s", st.PHPRC)
		ui.PrintLine("In PATH:   %v", st.InPath)
		if st.InPath && !st.PathAtFront {
			ui.PrintLine("PATH note:  PVM is not first in user PATH (index %d); run pvm setup or pvm refresh", st.PathIndex)
		}
		ui.PrintLine("")
		ui.PrintLine("Run 'pvm refresh' for session refresh commands.")
	},
}

func init() {
	rootCmd.AddCommand(pathCmd)
}
