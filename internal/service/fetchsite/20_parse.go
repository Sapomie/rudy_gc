package fetchsite

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var infoHashPattern = regexp.MustCompile(`(?i)btih:([0-9a-z]+)`)

func NormalizeMovieCode(movieCode string) string {
	code := strings.TrimSpace(movieCode)
	code = strings.ReplaceAll(code, "_", "-")
	return code
}

func BuildSukebeiQuery(movieCode string) string {
	code := strings.ToUpper(NormalizeMovieCode(movieCode))
	parts := strings.Split(code, "-")
	if len(parts) != 2 {
		return ""
	}
	return parts[0] + " " + parts[1]
}

func ParseSizeBytes(sizeText string) int64 {
	raw := strings.TrimSpace(strings.ToUpper(sizeText))
	raw = strings.ReplaceAll(raw, "IB", "B")
	raw = strings.ReplaceAll(raw, " ", "")
	if raw == "" {
		return 0
	}

	type unitDef struct {
		suffix string
		value  float64
	}
	units := []unitDef{
		{suffix: "TB", value: 1024 * 1024 * 1024 * 1024},
		{suffix: "GB", value: 1024 * 1024 * 1024},
		{suffix: "MB", value: 1024 * 1024},
		{suffix: "KB", value: 1024},
		{suffix: "B", value: 1},
	}
	for _, unit := range units {
		if !strings.HasSuffix(raw, unit.suffix) {
			continue
		}
		numText := strings.TrimSuffix(raw, unit.suffix)
		num, err := strconv.ParseFloat(numText, 64)
		if err != nil {
			return 0
		}
		return int64(num * unit.value)
	}
	return 0
}

func ParseDateTime(raw string) int64 {
	text := strings.TrimSpace(raw)
	if text == "" {
		return 0
	}

	layouts := []string{
		time.DateTime,
		"2006-01-02",
		"2006/01/02 15:04",
		"2006/01/02",
		"2006-01-02 15:04",
		time.RFC3339,
	}
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, text, time.Local)
		if err == nil {
			return t.Unix()
		}
	}

	if unixTs, err := strconv.ParseInt(text, 10, 64); err == nil {
		return unixTs
	}
	return 0
}

func ParseInfoHash(magnetURL string) string {
	matches := infoHashPattern.FindStringSubmatch(magnetURL)
	if len(matches) < 2 {
		return ""
	}
	return strings.ToUpper(matches[1])
}

func ParseDN(magnetURL string) string {
	parsed, err := url.Parse(magnetURL)
	if err != nil {
		return ""
	}
	return parsed.Query().Get("dn")
}

func ParseInt64(text string) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func RequireHTTP200(respStatus int) error {
	if respStatus != 200 {
		return fmt.Errorf("unexpected status: %d", respStatus)
	}
	return nil
}
