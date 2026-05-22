package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/php"
	"tagecode/pvm/pkg/ui"
)

var (
	listRemoteMajor string
	listRemoteMinor string
)

var listRemoteCmd = &cobra.Command{
	Use:     "list-remote",
	Aliases: []string{"ls-remote"},
	Short:   "List PHP versions available for download",
	Run: func(cmd *cobra.Command, args []string) {
		a, err := initApp()
		if err != nil {
			handleErr(cmd, app.SystemError("init failed", err.Error()))
			return
		}
		var major, minor *int
		if listRemoteMajor != "" {
			v, err := strconv.Atoi(listRemoteMajor)
			if err != nil {
				handleErr(cmd, app.UserError("invalid --major value", ""))
				return
			}
			major = &v
		}
		if listRemoteMinor != "" {
			parts := splitMinor(listRemoteMinor)
			if len(parts) != 2 {
				handleErr(cmd, app.UserError("invalid --minor value, use 8.3", ""))
				return
			}
			mj, _ := strconv.Atoi(parts[0])
			mn, _ := strconv.Atoi(parts[1])
			major = &mj
			minor = &mn
		}

		releases, err := a.Install.Remote.List(cmd.Context(), major, minor)
		if err != nil {
			handleErr(cmd, err)
			return
		}
		if ui.JSONEnabled(cmd) {
			_ = ui.PrintJSON(toListRemoteJSON(releases))
			return
		}
		if len(releases) == 0 {
			ui.PrintLine("No remote releases found.")
			return
		}
		ui.PrintTable(toListRemoteTable(releases))
	},
}

func toListRemoteTable(releases []php.RemoteRelease) ui.Table {
	rows := make([][]string, 0, len(releases))
	for _, r := range releases {
		rows = append(rows, []string{
			remoteVersionLabel(r.Version),
			r.Version.Arch,
			remoteTSLabel(r.Version.ThreadSafe),
			r.Version.VC,
			remoteSourceLabel(r.Archived),
			r.ZipFile,
		})
	}
	return ui.Table{
		Headers: []string{"VERSION", "ARCH", "TS", "VC", "SOURCE", "PACKAGE"},
		Rows:    rows,
	}
}

func toListRemoteJSON(releases []php.RemoteRelease) map[string]any {
	type row struct {
		ID         string `json:"id"`
		Version    string `json:"version"`
		Arch       string `json:"arch"`
		ThreadSafe bool   `json:"thread_safe"`
		VC         string `json:"vc"`
		Source     string `json:"source"`
		ZipFile    string `json:"zip_file"`
		URL        string `json:"url"`
		Archived   bool   `json:"archived"`
	}
	out := make([]row, 0, len(releases))
	for _, r := range releases {
		out = append(out, row{
			ID:         r.Version.ID(),
			Version:    remoteVersionLabel(r.Version),
			Arch:       r.Version.Arch,
			ThreadSafe: r.Version.ThreadSafe,
			VC:         r.Version.VC,
			Source:     remoteSourceLabel(r.Archived),
			ZipFile:    r.ZipFile,
			URL:        r.URL,
			Archived:   r.Archived,
		})
	}
	return map[string]any{"releases": out}
}

func remoteVersionLabel(v php.Version) string {
	label := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Revision > 0 {
		label += fmt.Sprintf("-%d", v.Revision)
	} else if v.Prerelease != "" {
		label += "-" + v.Prerelease
	}
	return label
}

func remoteTSLabel(threadSafe bool) string {
	if threadSafe {
		return "ts"
	}
	return "nts"
}

func remoteSourceLabel(archived bool) string {
	if archived {
		return "archives"
	}
	return "releases"
}

func splitMinor(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}

func init() {
	listRemoteCmd.Flags().StringVar(&listRemoteMajor, "major", "", "Filter by major version")
	listRemoteCmd.Flags().StringVar(&listRemoteMinor, "minor", "", "Filter by major.minor")
	rootCmd.AddCommand(listRemoteCmd)
}
