package html

import "strings"

func singleActorFilterName(raw string) string {
	normalized := raw
	for _, sep := range []string{",", "，", "、", "/", "|"} {
		normalized = strings.ReplaceAll(normalized, sep, " ")
	}

	parts := strings.Fields(normalized)
	if len(parts) != 1 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}
