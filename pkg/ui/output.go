package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var stdout io.Writer = os.Stdout
var stderr io.Writer = os.Stderr

func SetOutput(out, errOut io.Writer) {
	stdout = out
	stderr = errOut
}

func JSONEnabled(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	v, _ := cmd.Root().PersistentFlags().GetBool("json")
	return v
}

func QuietEnabled(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	v, _ := cmd.Root().PersistentFlags().GetBool("quiet")
	return v
}

func PrintJSON(v any) error {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func PrintJSONError(message, hint string) error {
	return PrintJSON(map[string]string{
		"error": message,
		"hint":  hint,
	})
}

func PrintLine(format string, args ...any) {
	fmt.Fprintf(stdout, format+"\n", args...)
}

func PrintError(err error) {
	fmt.Fprintf(stderr, "error: %v\n", err)
}

func PrintHint(hint string) {
	if hint != "" {
		fmt.Fprintf(stderr, "hint: %s\n", hint)
	}
}
