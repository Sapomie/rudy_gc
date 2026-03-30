package movie

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

func (s *Service) buildMovieFetchSiteDetail(ctx context.Context, javID string) (*types.MovieFetchSiteStatus, []*types.MovieJavbusMagnet, *types.MovieFetchSiteStatus, []*types.MovieSukebeiTorrent, error) {
	javbusFetch, err := s.buildJavbusFetchStatus(ctx, javID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	javbusMagnets, err := s.buildJavbusMagnets(ctx, javID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	sukebeiFetch, err := s.buildSukebeiFetchStatus(ctx, javID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	sukebeiTorrents, err := s.buildSukebeiTorrents(ctx, javID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if err := s.applyFetchSiteFavoriteStatus(ctx, javID, javbusMagnets, sukebeiTorrents); err != nil {
		return nil, nil, nil, nil, err
	}
	applyMatchedFetchSiteHashes(javbusMagnets, sukebeiTorrents)
	return javbusFetch, javbusMagnets, sukebeiFetch, sukebeiTorrents, nil
}

func (s *Service) buildJavbusFetchStatus(ctx context.Context, javID string) (*types.MovieFetchSiteStatus, error) {
	row, err := s.deps.JavbusMagnetFetchModel.FindOneByMovieJavId(ctx, javID)
	if err == moviex.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &types.MovieFetchSiteStatus{
		MovieJavID:        row.MovieJavId,
		MovieName:         row.MovieName,
		ReleaseDate:       tsToDate(row.ReleaseDate),
		FetchStatus:       row.FetchStatus,
		FetchStatusText:   fetchStatusText(row.FetchStatus),
		TryCount:          row.TryCount,
		LastFetchTime:     tsToDateTime(row.LastFetchTime),
		LastFetchAgo:      tsToDaysAgo(row.LastFetchTime),
		LastError:         strings.TrimSpace(row.LastError),
		TorrentHashCount:  row.TorrentHashCount,
		LatestPublishTime: tsToDate(row.LatestPublishTime),
		SourceURL:         row.SourceUrl,
	}, nil
}

func (s *Service) buildJavbusMagnets(ctx context.Context, javID string) ([]*types.MovieJavbusMagnet, error) {
	rows, err := s.deps.JavbusMagnetModel.ListByMovieJavId(ctx, javID)
	if err != nil {
		return nil, err
	}
	out := make([]*types.MovieJavbusMagnet, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &types.MovieJavbusMagnet{
			RowID:       row.Id,
			MagnetName:  row.MagnetName,
			MagnetURL:   row.MagnetUrl,
			InfoHash:    row.InfoHash,
			SizeText:    row.SizeText,
			ShareDate:   tsToDate(row.ShareDate),
			HasHD:       row.HasHd == 1,
			HasSubtitle: row.HasSubtitle == 1,
			PageURL:     row.PageUrl,
		})
	}
	return out, nil
}

func (s *Service) buildSukebeiFetchStatus(ctx context.Context, javID string) (*types.MovieFetchSiteStatus, error) {
	row, err := s.deps.SukebeiTorrentFetchModel.FindOneByMovieJavId(ctx, javID)
	if err == moviex.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &types.MovieFetchSiteStatus{
		MovieJavID:        row.MovieJavId,
		MovieName:         row.MovieName,
		ReleaseDate:       tsToDate(row.ReleaseDate),
		FetchStatus:       row.FetchStatus,
		FetchStatusText:   fetchStatusText(row.FetchStatus),
		TryCount:          row.TryCount,
		LastFetchTime:     tsToDateTime(row.LastFetchTime),
		LastFetchAgo:      tsToDaysAgo(row.LastFetchTime),
		LastError:         strings.TrimSpace(row.LastError),
		TorrentHashCount:  row.TorrentHashCount,
		LatestPublishTime: tsToDate(row.LatestPublishTime),
		SourceURL:         row.SourceUrl,
	}, nil
}

func (s *Service) buildSukebeiTorrents(ctx context.Context, javID string) ([]*types.MovieSukebeiTorrent, error) {
	rows, err := s.deps.SukebeiTorrentModel.ListByMovieJavId(ctx, javID)
	if err != nil {
		return nil, err
	}
	out := make([]*types.MovieSukebeiTorrent, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &types.MovieSukebeiTorrent{
			RowID:        row.Id,
			TorrentTitle: row.TorrentTitle,
			ViewURL:      row.ViewUrl,
			TorrentURL:   row.TorrentUrl,
			MagnetURL:    row.MagnetUrl,
			InfoHash:     row.InfoHash,
			SizeText:     row.SizeText,
			PublishTime:  tsToDate(row.PublishTime),
			Seeders:      row.Seeders,
			Leechers:     row.Leechers,
			Completed:    row.Completed,
		})
	}
	return out, nil
}

func fetchStatusText(status int64) string {
	switch status {
	case 1:
		return "待抓取"
	case 2:
		return "抓取中"
	case 3:
		return "成功"
	case 4:
		return "失败"
	default:
		return "未入队"
	}
}

func tsToDateTime(ts int64) string {
	if ts <= 0 {
		return "-"
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

func tsToDaysAgo(ts int64) string {
	if ts <= 0 {
		return "-"
	}
	days := time.Since(time.Unix(ts, 0)).Hours() / 24
	return fmt.Sprintf("%.1f 天", days)
}

func applyMatchedFetchSiteHashes(javbusMagnets []*types.MovieJavbusMagnet, sukebeiTorrents []*types.MovieSukebeiTorrent) {
	if len(javbusMagnets) == 0 || len(sukebeiTorrents) == 0 {
		return
	}

	javbusHashes := make(map[string]struct{}, len(javbusMagnets))
	sukebeiHashes := make(map[string]struct{}, len(sukebeiTorrents))

	for _, row := range javbusMagnets {
		if row == nil {
			continue
		}
		hash := normalizeInfoHash(row.InfoHash)
		if hash == "" {
			continue
		}
		javbusHashes[hash] = struct{}{}
	}

	for _, row := range sukebeiTorrents {
		if row == nil {
			continue
		}
		hash := normalizeInfoHash(row.InfoHash)
		if hash == "" {
			continue
		}
		sukebeiHashes[hash] = struct{}{}
	}

	for _, row := range javbusMagnets {
		if row == nil {
			continue
		}
		_, ok := sukebeiHashes[normalizeInfoHash(row.InfoHash)]
		row.HasMatchedHash = ok
	}

	for _, row := range sukebeiTorrents {
		if row == nil {
			continue
		}
		_, ok := javbusHashes[normalizeInfoHash(row.InfoHash)]
		row.HasMatchedHash = ok
	}
}

func (s *Service) applyFetchSiteFavoriteStatus(ctx context.Context, javID string, javbusMagnets []*types.MovieJavbusMagnet, sukebeiTorrents []*types.MovieSukebeiTorrent) error {
	if len(javbusMagnets) == 0 && len(sukebeiTorrents) == 0 {
		return nil
	}

	items, err := s.ListDefaultAlbumItemsByMovieJavID(ctx, javID)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	favorited := make(map[string]struct{}, len(items))
	for _, row := range items {
		if row == nil {
			continue
		}
		favorited[favoriteKey(row.SourceType, row.SourceRowId)] = struct{}{}
	}

	for _, row := range javbusMagnets {
		if row == nil {
			continue
		}
		_, ok := favorited[favoriteKey(sourceTypeJavbusMagnet, row.RowID)]
		row.IsFavorited = ok
	}
	for _, row := range sukebeiTorrents {
		if row == nil {
			continue
		}
		_, ok := favorited[favoriteKey(sourceTypeSukebei, row.RowID)]
		row.IsFavorited = ok
	}
	return nil
}

func favoriteKey(sourceType string, sourceRowID int64) string {
	return strings.TrimSpace(sourceType) + ":" + fmt.Sprintf("%d", sourceRowID)
}

func normalizeInfoHash(hash string) string {
	return strings.ToUpper(strings.TrimSpace(hash))
}
