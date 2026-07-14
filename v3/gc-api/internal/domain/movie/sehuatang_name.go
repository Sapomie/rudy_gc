package movie

import (
	"regexp"
	"strings"
)

var sehuatangJAVIDPattern = regexp.MustCompile(`(?i)\b([a-z0-9]{2,12}-\d{2,7})\b`)

// ParseSehuatangMovieName extracts the normalized movie name used by the
// Sehuatang fetch and list flows.
func ParseSehuatangMovieName(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}
	match := sehuatangJAVIDPattern.FindStringSubmatch(title)
	if len(match) < 2 {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(match[1]))
}
