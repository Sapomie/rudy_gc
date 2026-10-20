package sc

import (
	"path/filepath"
	"strings"
	"unicode"

	"rudy_gc/local/api/internal/types"
)

const smartPickNameFallback = "NA"

func (l *Service) smartPickCopyFileName(movie *types.MovieType, srcFilePath, source string) string {
	srcName := strings.TrimSpace(filepath.Base(srcFilePath))
	if srcName == "" {
		return ""
	}

	if NormalizeSmartPickSource(source) != SmartPickSourceWMedia {
		return srcName
	}

	ext := filepath.Ext(srcName)
	movieName := cleanNameToken(movieNameToken(movie))
	castName := cleanNameToken(castNameToken(movie))
	title := cleanNameToken(titleToken(movie))

	if movieName == "" {
		movieName = smartPickNameFallback
	}
	if castName == "" {
		castName = smartPickNameFallback
	}
	if title == "" {
		title = smartPickNameFallback
	}

	return movieName + "_" + castName + "_" + title + ext
}

func smartPickFC2CopyFileName(movie *FC2SmartPickMovie, srcFilePath string) string {
	srcName := strings.TrimSpace(filepath.Base(srcFilePath))
	if srcName == "" {
		return ""
	}
	ext := filepath.Ext(srcName)
	movieName := ""
	performer := ""
	title := ""
	if movie != nil && movie.FC2Movie != nil {
		movieName = cleanNameToken(movie.FC2Key)
		performer = cleanNameToken(movie.PerformerText)
		title = cleanNameToken(movie.Title)
	}
	if movieName == "" {
		movieName = smartPickNameFallback
	}
	if performer == "" {
		performer = smartPickNameFallback
	}
	if title == "" {
		title = smartPickNameFallback
	}
	return movieName + "_" + performer + "_" + title + ext
}

func movieNameToken(movie *types.MovieType) string {
	if movie == nil {
		return ""
	}
	return strings.TrimSpace(movie.Name)
}

func castNameToken(movie *types.MovieType) string {
	if movie == nil || len(movie.Cast) == 0 {
		return ""
	}

	names := make([]string, 0, len(movie.Cast))
	seen := make(map[string]struct{}, len(movie.Cast))
	for _, cast := range movie.Cast {
		if cast == nil {
			continue
		}
		name := strings.TrimSpace(cast.NameShow)
		if name == "" {
			name = strings.TrimSpace(cast.DisplayName)
		}
		if name == "" {
			name = strings.TrimSpace(cast.Name)
		}
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return strings.Join(names, "-")
}

func titleToken(movie *types.MovieType) string {
	if movie == nil {
		return ""
	}
	return strings.TrimSpace(movie.Title)
}

func cleanNameToken(s string) string {
	if s == "" {
		return ""
	}

	s = strings.Join(strings.Fields(strings.TrimSpace(s)), "")
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsControl(r) {
			continue
		}
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			continue
		default:
			b.WriteRune(r)
		}
	}

	cleaned := strings.TrimSpace(b.String())
	cleaned = strings.Trim(cleaned, ".")
	return cleaned
}
