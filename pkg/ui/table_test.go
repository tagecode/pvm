package ui

import "testing"

func TestTableRender(t *testing.T) {
	got := Table{
		Headers: []string{"VERSION", "ARCH", "TS"},
		Rows: [][]string{
			{"8.3.31", "x64", "nts"},
			{"7.4.33", "x86", "ts"},
		},
	}.Render()

	want := "" +
		"VERSION  ARCH  TS \n" +
		"8.3.31   x64   nts\n" +
		"7.4.33   x86   ts \n"

	if got != want {
		t.Fatalf("Render() =\n%q\nwant\n%q", got, want)
	}
}

func TestTableRenderEmptyRows(t *testing.T) {
	got := Table{
		Headers: []string{"VERSION", "ARCH"},
		Rows:    nil,
	}.Render()

	want := "VERSION  ARCH\n"
	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}
