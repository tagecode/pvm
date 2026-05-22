package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall <version>",
	Aliases: []string{"remove"},
	Short:   "Uninstall a PHP version",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		yes, _ := cmd.Root().PersistentFlags().GetBool("yes")
		if !yes && !dryRun(cmd) {
			fmt.Printf("Uninstall %s? [y/N]: ", args[0])
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				return
			}
		}
		if err := a.Install.Uninstall(cmd.Context(), args[0]); err != nil {
			handleErr(cmd, err)
			return
		}
		if !dryRun(cmd) {
			fmt.Printf("Uninstalled %s\n", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
