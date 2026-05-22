package ui

import (
	"fmt"
	"strings"
)

// Table renders fixed-width text columns for CLI output.
type Table struct {
	Headers []string
	Rows    [][]string
}

func PrintTable(t Table) {
	if len(t.Headers) == 0 {
		return
	}
	fmt.Fprint(stdout, t.Render())
}

func (t Table) Render() string {
	if len(t.Headers) == 0 {
		return ""
	}

	colCount := len(t.Headers)
	widths := make([]int, colCount)
	for i, h := range t.Headers {
		widths[i] = len(h)
	}
	for _, row := range t.Rows {
		for i := 0; i < colCount && i < len(row); i++ {
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	var b strings.Builder
	writeRow := func(cells []string) {
		for i := 0; i < colCount; i++ {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(cell)
			if pad := widths[i] - len(cell); pad > 0 {
				b.WriteString(strings.Repeat(" ", pad))
			}
		}
		b.WriteString("\n")
	}

	writeRow(t.Headers)
	for _, row := range t.Rows {
		writeRow(row)
	}
	return b.String()
}
