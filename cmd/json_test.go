package cmd

import (
	"encoding/json"
	"testing"

	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/php"
)

func TestToListJSON(t *testing.T) {
	items := []app.InstalledVersion{{
		ID:      "8.3.31-x64-nts-vs16",
		Version: php.Version{Major: 8, Minor: 3, Patch: 31, Arch: "x64", ThreadSafe: false, VC: "vs16"},
		Path:    `C:\pvm\versions\8.3.31-x64-nts-vs16`,
		Active:  true,
	}}
	out := toListJSON(items)
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatalf("invalid json: %s", data)
	}
}
