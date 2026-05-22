package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pvm",
	Short: "PHP Version Manager for Windows",
	Long: `PVM manages multiple PHP versions on Windows.

Common commands:
  pvm install 8.3          Install latest PHP 8.3.x
  pvm use 8.3.31           Switch active PHP version
  pvm list                 List installed versions
  pvm setup                Manually configure user PATH (auto on first use)
  pvm refresh              Print session PATH refresh commands

Run 'pvm <command> --help' for details.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet mode")
	rootCmd.PersistentFlags().CountP("verbose", "v", "Verbose logging")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Simulate operations without making changes")
}
