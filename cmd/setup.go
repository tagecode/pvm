package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/ui"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure user PATH for PVM",
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		system, _ := cmd.Flags().GetBool("system")
		if err := a.Setup.Run(cmd.Context(), app.SetupOptions{System: system, DryRun: dryRun(cmd)}); err != nil {
			handleErr(cmd, err)
			return
		}
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(map[string]string{"status": "ok"})
			return
		}
		if dryRun(cmd) {
			ui.PrintLine("Would configure user PATH and PHP_HOME")
			return
		}
		ui.PrintLine("PVM setup complete. User PATH and PHP_HOME are configured.")
	},
}

var unsetupCmd = &cobra.Command{
	Use:   "unsetup",
	Short: "Remove PVM entries from user environment variables",
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		system, _ := cmd.Flags().GetBool("system")
		if err := a.Setup.Unsetup(cmd.Context(), app.SetupOptions{System: system, DryRun: dryRun(cmd)}); err != nil {
			handleErr(cmd, err)
			return
		}
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(map[string]string{"status": "ok"})
			return
		}
		if dryRun(cmd) {
			ui.PrintLine("Would remove PVM from user PATH and PHP_HOME")
			return
		}
		ui.PrintLine("PVM environment entries removed.")
	},
}

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Print commands to refresh PATH in the current session",
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		ps, c := a.Setup.RefreshScript()
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(map[string]string{"powershell": ps, "cmd": c})
			return
		}
		ui.PrintLine("Refresh current session (PowerShell):")
		ui.PrintLine("  %s", ps)
		ui.PrintLine("Refresh current session (cmd):")
		ui.PrintLine("  %s", c)
	},
}

func init() {
	setupCmd.Flags().Bool("system", false, "Configure system-level environment (requires admin)")
	unsetupCmd.Flags().Bool("system", false, "Remove system-level environment (requires admin)")
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(unsetupCmd)
	rootCmd.AddCommand(refreshCmd)
}
