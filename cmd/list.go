package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/ui"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List installed PHP versions",
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		items, err := a.Registry.List()
		if err != nil {
			handleErr(cmd, err)
			return
		}
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(toListJSON(items))
			return
		}
		if len(items) == 0 {
			ui.PrintLine("No PHP versions installed. Run 'pvm install <version>'.")
			return
		}
		for _, v := range items {
			markers := formatMarkers(v.Active, v.Default)
			ui.PrintLine("%s %s  %s", v.ID, markers, v.Path)
		}
	},
}

func formatMarkers(active, defaultVer bool) string {
	var parts []string
	if active {
		parts = append(parts, "*")
	}
	if defaultVer {
		parts = append(parts, "(default)")
	}
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += " " + p
	}
	return out
}

func init() {
	rootCmd.AddCommand(listCmd)
}
