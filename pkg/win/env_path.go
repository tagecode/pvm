package win

import "strings"

// PrependPathValue inserts entry at the front of a semicolon-separated PATH string.
// Duplicate entries (case-insensitive) are removed before prepending.
func PrependPathValue(path, entry string) string {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return path
	}

	var kept []string
	for _, p := range strings.Split(path, ";") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.EqualFold(p, entry) {
			continue
		}
		kept = append(kept, p)
	}
	if len(kept) == 0 {
		return entry
	}
	return entry + ";" + strings.Join(kept, ";")
}
