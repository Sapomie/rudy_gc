package movie

import (
	"strings"

	fc2domain "rudy_gc/local/api/internal/types/fc2"
)

const (
	SehuatangThreadMovieTypeJAV    = "jav"
	SehuatangThreadMovieTypeFC2PPV = "fc2ppv"
)

func ResolveSehuatangThreadMovieType(movieName, movieJavID string) string {
	if _, _, ok := fc2domain.NormalizeCode(movieName); ok {
		return SehuatangThreadMovieTypeFC2PPV
	}
	if strings.TrimSpace(movieJavID) != "" {
		return SehuatangThreadMovieTypeJAV
	}
	return ""
}

func NormalizeSehuatangThreadMovieTypes(values ...string) []string {
	types := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != SehuatangThreadMovieTypeJAV && value != SehuatangThreadMovieTypeFC2PPV {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		types = append(types, value)
	}
	return types
}
