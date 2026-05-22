package cmd

import (
	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/php"
	"tagecode/pvm/pkg/ui"
)

var (
	installArch      string
	installTS        bool
	installNTS       bool
	installVC        string
	installFromZip   string
	installReinstall bool
)

var installCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Download and install a PHP version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		spec, err := php.ParseSpec(args[0])
		if err != nil {
			handleErr(cmd, app.UserError(err.Error(), "example: pvm install 8.3"))
			return
		}

		opts := app.InstallOptions{
			Arch:      installArch,
			VC:        installVC,
			FromZip:   installFromZip,
			Reinstall: installReinstall,
			DryRun:    dryRun(cmd),
		}
		if installTS {
			v := true
			opts.ThreadSafe = &v
		} else if installNTS {
			v := false
			opts.ThreadSafe = &v
		}

		iv, err := a.Install.Install(cmd.Context(), spec, opts)
		if err != nil {
			handleErr(cmd, err)
			return
		}
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(iv)
			return
		}
		if dryRun(cmd) {
			ui.PrintLine("Would install %s to %s", iv.ID, iv.Path)
			return
		}
		if iv.Active {
			ui.PrintLine("Installed %s at %s (now active, php ready in this terminal)", iv.ID, iv.Path)
			return
		}
		ui.PrintLine("Installed %s at %s", iv.ID, iv.Path)
	},
}

func init() {
	installCmd.Flags().StringVar(&installArch, "arch", "", "Architecture: x64, x86, or auto")
	installCmd.Flags().BoolVar(&installTS, "ts", false, "Install Thread Safe build")
	installCmd.Flags().BoolVar(&installNTS, "nts", false, "Install Non-Thread Safe build")
	installCmd.Flags().StringVar(&installVC, "vc", "", "VC runtime: vs16, vs17, vc15")
	installCmd.Flags().StringVar(&installFromZip, "from-zip", "", "Install from local zip file")
	installCmd.Flags().BoolVar(&installReinstall, "reinstall", false, "Force reinstall")
	rootCmd.AddCommand(installCmd)
}
