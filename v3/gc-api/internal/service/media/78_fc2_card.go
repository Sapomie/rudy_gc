package media

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
)

// FC2CardView is the only response shape used by every FC2 movie-card list.
type FC2CardView struct {
	*moviex.FC2Movie
	EffectiveCoverImageID int64
	CoverImageURL         string
	VideoURL              string
	PlayButtonClass       string
	PlayButtonText        string
	ShowPlayButton        bool
	DownloadTime          int64
	ScTimes               int64
	ComeTimes             int64
	LastScTime            int64
}

// BuildFC2CardViews loads card data with separate batch queries and preserves
// the input movie order. It deliberately performs no JOIN queries.
func (s *Service) BuildFC2CardViews(ctx context.Context, movieRows []*moviex.FC2Movie) ([]*FC2CardView, error) {
	if s == nil || s.deps == nil {
		return nil, fmt.Errorf("media service dependencies are nil")
	}
	movieIDs := make([]int64, 0, len(movieRows))
	for _, row := range movieRows {
		if row != nil && row.ID > 0 {
			movieIDs = append(movieIDs, row.ID)
		}
	}
	mediaRows, err := s.deps.FC2MediaModel.ListByMovieIDs(ctx, movieIDs)
	if err != nil {
		return nil, fmt.Errorf("FC2MediaModel.ListByMovieIDs: %w", err)
	}
	statRows, err := s.deps.FC2ScStatModel.ListByMovieIDs(ctx, movieIDs)
	if err != nil {
		return nil, fmt.Errorf("FC2ScStatModel.ListByMovieIDs: %w", err)
	}
	mediaIDs := make([]int64, 0, len(mediaRows))
	for _, row := range mediaRows {
		if row != nil && row.ID > 0 {
			mediaIDs = append(mediaIDs, row.ID)
		}
	}
	imageRows, err := s.deps.FC2MediaImageModel.ListByMediaIDs(ctx, mediaIDs)
	if err != nil {
		return nil, fmt.Errorf("FC2MediaImageModel.ListByMediaIDs: %w", err)
	}
	return AssembleFC2CardViews(movieRows, mediaRows, statRows, imageRows), nil
}

// AssembleFC2CardViews applies the shared card projection to preloaded rows.
func AssembleFC2CardViews(movieRows []*moviex.FC2Movie, mediaRows []*moviex.FC2Media, statRows []*moviex.FC2ScStat, imageRows []*moviex.FC2MediaImage) []*FC2CardView {
	mediaByMovieID := make(map[int64]*moviex.FC2Media, len(mediaRows))
	for _, row := range mediaRows {
		if row != nil && row.FC2MovieID > 0 {
			mediaByMovieID[row.FC2MovieID] = row
		}
	}
	statByMovieID := make(map[int64]*moviex.FC2ScStat, len(statRows))
	for _, row := range statRows {
		if row != nil && row.FC2MovieID > 0 {
			statByMovieID[row.FC2MovieID] = row
		}
	}
	imagesByMediaID := make(map[int64][]*moviex.FC2MediaImage, len(imageRows))
	for _, row := range imageRows {
		if row != nil && row.FC2MediaID > 0 {
			imagesByMediaID[row.FC2MediaID] = append(imagesByMediaID[row.FC2MediaID], row)
		}
	}

	items := make([]*FC2CardView, 0, len(movieRows))
	for _, movieRow := range movieRows {
		if movieRow == nil {
			continue
		}
		item := &FC2CardView{FC2Movie: movieRow}
		if mediaRow := mediaByMovieID[movieRow.ID]; mediaRow != nil {
			applyFC2CardMedia(item, mediaRow)
			if imageRow := ResolveFC2CoverImage(mediaRow.ID, movieRow.CoverImageID, imagesByMediaID[mediaRow.ID]); imageRow != nil {
				item.EffectiveCoverImageID = imageRow.ID
				item.CoverImageURL = filepath.ToSlash(strings.TrimSpace(imageRow.FilePath))
			}
		}
		if statRow := statByMovieID[movieRow.ID]; statRow != nil {
			item.ScTimes = statRow.ScTimes
			item.ComeTimes = statRow.ComeTimes
			item.LastScTime = statRow.LastScTime
		}
		items = append(items, item)
	}
	return items
}

func applyFC2CardMedia(item *FC2CardView, mediaRow *moviex.FC2Media) {
	if item == nil || mediaRow == nil {
		return
	}
	item.DownloadTime = mediaRow.BirthTime
	if strings.TrimSpace(mediaRow.FullDir) == "" || strings.TrimSpace(mediaRow.FileName) == "" {
		return
	}
	item.VideoURL = filepath.Join(mediaRow.FullDir, mediaRow.FileName)
	item.PlayButtonText = "Play"
	item.ShowPlayButton = true
	switch {
	case mediaRow.IsRemoved == int64(consts.FilmIsRemoved):
		item.PlayButtonClass = "btn-secondary"
		item.PlayButtonText = "Removed"
	case mediaRow.HasSub == int64(consts.FilmHasSub):
		item.PlayButtonClass = "btn-success"
	default:
		item.PlayButtonClass = "btn-info"
	}
}
