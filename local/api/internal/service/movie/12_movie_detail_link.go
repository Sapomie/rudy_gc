package movie

import (
	"context"
	"net/url"
	"strings"

	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type MovieDetailLinkRef struct {
	MovieJavID string
	MovieName  string
}

func (s *Service) ResolveMovieDetailHrefs(ctx context.Context, refs []MovieDetailLinkRef) ([]string, error) {
	if len(refs) == 0 {
		return []string{}, nil
	}

	javIDs := make([]string, 0, len(refs))
	names := make([]string, 0, len(refs))
	for _, ref := range refs {
		if javID := strings.TrimSpace(ref.MovieJavID); javID != "" {
			javIDs = append(javIDs, javID)
		}
		if name := strings.TrimSpace(ref.MovieName); name != "" {
			names = append(names, name)
		}
	}
	rows, err := s.deps.MovieModel.FindMovieIdentities(ctx, javIDs, names)
	if err != nil {
		return nil, err
	}
	return buildMovieDetailHrefs(refs, rows), nil
}

func buildMovieDetailHrefs(refs []MovieDetailLinkRef, rows []*moviex.AMovieIdentity) []string {
	nameByJavID := make(map[string]string, len(rows))
	existingNames := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		javID := strings.TrimSpace(row.JavID)
		name := strings.TrimSpace(row.Name)
		if javID != "" && name != "" {
			nameByJavID[javID] = name
		}
		if name != "" {
			existingNames[name] = struct{}{}
		}
	}

	hrefs := make([]string, len(refs))
	for index, ref := range refs {
		javID := strings.TrimSpace(ref.MovieJavID)
		name := strings.TrimSpace(ref.MovieName)
		if canonicalName := nameByJavID[javID]; canonicalName != "" {
			hrefs[index] = "/moviedetail/" + url.PathEscape(canonicalName)
			continue
		}
		if _, ok := existingNames[name]; ok && name != "" {
			hrefs[index] = "/moviedetail/" + url.PathEscape(name)
		}
	}
	return hrefs
}
