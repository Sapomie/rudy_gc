package movie

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

func (s *Service) buildMovieFetchSiteDetail(ctx context.Context, javID string, movieName string) (*types.MovieFetchSiteStatus, []*types.MovieJavbusMagnet, *types.MovieFetchSiteStatus, []*types.MovieSukebeiTorrent, []*types.MovieSehuatangMagnet, error) {
	javbusFetch, err := s.buildJavbusFetchStatus(ctx, javID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	javbusMagnets, err := s.buildJavbusMagnets(ctx, javID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	sukebeiFetch, err := s.buildSukebeiFetchStatus(ctx, javID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	sukebeiTorrents, err := s.buildSukebeiTorrents(ctx, javID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	sehuatangMagnets, err := s.buildSehuatangMagnets(ctx, javID, movieName)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err := s.applyFetchSiteFavoriteStatus(ctx, javID, javbusMagnets, sukebeiTorrents, sehuatangMagnets); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	applyMatchedFetchSiteHashes(javbusMagnets, sukebeiTorrents, sehuatangMagnets)
	return javbusFetch, javbusMagnets, sukebeiFetch, sukebeiTorrents, sehuatangMagnets, nil
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
			InfoHash:    row.InfoHash,
			SizeText:    formatResourceSizeBytes(row.SizeBytes),
			ShareDate:   tsToDate(row.ShareDate),
			HasHD:       row.HasHd == 1,
			HasSubtitle: row.HasSubtitle == 1,
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
			ViewURL:      buildSukebeiViewURL(row.ViewId),
			InfoHash:     row.InfoHash,
			SizeText:     formatResourceSizeBytes(row.SizeBytes),
			PublishTime:  tsToDate(row.PublishTime),
			Seeders:      row.Seeders,
			Leechers:     row.Leechers,
			Completed:    row.Completed,
		})
	}
	return out, nil
}

func (s *Service) buildSehuatangMagnets(ctx context.Context, javID string, movieName string) ([]*types.MovieSehuatangMagnet, error) {
	rows, err := s.deps.SehuatangMagnetModel.ListByMovieKey(ctx, javID, movieName)
	if err != nil {
		return nil, err
	}
	out := make([]*types.MovieSehuatangMagnet, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &types.MovieSehuatangMagnet{
			RowID:       row.Id,
			ThreadTitle: row.ThreadTitle,
			ThreadURL:   row.ThreadUrl,
			InfoHash:    row.InfoHash,
			PostTime:    tsToDateTime(row.PostTime),
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

func buildSukebeiViewURL(viewID int64) string {
	if viewID <= 0 {
		return ""
	}
	return fmt.Sprintf("https://sukebei.nyaa.si/view/%d", viewID)
}

func formatResourceSizeBytes(sizeBytes int64) string {
	if sizeBytes <= 0 {
		return "-"
	}

	type unitDef struct {
		label string
		value float64
	}
	units := []unitDef{
		{label: "TB", value: 1024 * 1024 * 1024 * 1024},
		{label: "GB", value: 1024 * 1024 * 1024},
		{label: "MB", value: 1024 * 1024},
		{label: "KB", value: 1024},
	}
	size := float64(sizeBytes)
	for _, unit := range units {
		if size >= unit.value {
			return fmt.Sprintf("%.2f %s", size/unit.value, unit.label)
		}
	}
	return fmt.Sprintf("%d B", sizeBytes)
}

func applyMatchedFetchSiteHashes(javbusMagnets []*types.MovieJavbusMagnet, sukebeiTorrents []*types.MovieSukebeiTorrent, sehuatangMagnets []*types.MovieSehuatangMagnet) {
	if len(javbusMagnets) == 0 && len(sukebeiTorrents) == 0 && len(sehuatangMagnets) == 0 {
		return
	}

	javbusHashes := make(map[string]struct{}, len(javbusMagnets))
	sukebeiHashes := make(map[string]struct{}, len(sukebeiTorrents))
	sehuatangHashes := make(map[string]struct{}, len(sehuatangMagnets))

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
	for _, row := range sehuatangMagnets {
		if row == nil {
			continue
		}
		hash := normalizeInfoHash(row.InfoHash)
		if hash == "" {
			continue
		}
		sehuatangHashes[hash] = struct{}{}
	}

	for _, row := range javbusMagnets {
		if row == nil {
			continue
		}
		hash := normalizeInfoHash(row.InfoHash)
		_, matchedSukebei := sukebeiHashes[hash]
		_, matchedSehuatang := sehuatangHashes[hash]
		row.HasMatchedHash = matchedSukebei || matchedSehuatang
	}

	for _, row := range sukebeiTorrents {
		if row == nil {
			continue
		}
		hash := normalizeInfoHash(row.InfoHash)
		_, matchedJavbus := javbusHashes[hash]
		_, matchedSehuatang := sehuatangHashes[hash]
		row.HasMatchedHash = matchedJavbus || matchedSehuatang
	}

	for _, row := range sehuatangMagnets {
		if row == nil {
			continue
		}
		hash := normalizeInfoHash(row.InfoHash)
		_, matchedJavbus := javbusHashes[hash]
		_, matchedSukebei := sukebeiHashes[hash]
		row.HasMatchedHash = matchedJavbus || matchedSukebei
	}
}

func (s *Service) applyFetchSiteFavoriteStatus(ctx context.Context, javID string, javbusMagnets []*types.MovieJavbusMagnet, sukebeiTorrents []*types.MovieSukebeiTorrent, sehuatangMagnets []*types.MovieSehuatangMagnet) error {
	if len(javbusMagnets) == 0 && len(sukebeiTorrents) == 0 && len(sehuatangMagnets) == 0 {
		return nil
	}

	downloadItems, _, err := s.ListAlbumItemsByMovieJavID(ctx, defaultAlbumName, javID)
	if err != nil {
		return err
	}
	pendingItems, _, err := s.ListAlbumItemsByMovieJavID(ctx, pendingAlbumName, javID)
	if err != nil {
		return err
	}

	inDownload := make(map[string]struct{}, len(downloadItems))
	for _, row := range downloadItems {
		if row == nil {
			continue
		}
		inDownload[favoriteKey(row.SourceType, row.SourceRowId)] = struct{}{}
	}

	inPending := make(map[string]struct{}, len(pendingItems))
	for _, row := range pendingItems {
		if row == nil {
			continue
		}
		inPending[favoriteKey(row.SourceType, row.SourceRowId)] = struct{}{}
	}

	for _, row := range javbusMagnets {
		if row == nil {
			continue
		}
		_, okDownload := inDownload[favoriteKey(sourceTypeJavbusMagnet, row.RowID)]
		_, okPending := inPending[favoriteKey(sourceTypeJavbusMagnet, row.RowID)]
		row.IsFavorited = okDownload // 兼容旧字段：IsFavorited 等价于下载中状态
		row.IsInDownload = okDownload
		row.IsInPending = okPending
	}
	for _, row := range sukebeiTorrents {
		if row == nil {
			continue
		}
		_, okDownload := inDownload[favoriteKey(sourceTypeSukebei, row.RowID)]
		_, okPending := inPending[favoriteKey(sourceTypeSukebei, row.RowID)]
		row.IsFavorited = okDownload
		row.IsInDownload = okDownload
		row.IsInPending = okPending
	}
	for _, row := range sehuatangMagnets {
		if row == nil {
			continue
		}
		_, okDownload := inDownload[favoriteKey(sourceTypeSehuatang, row.RowID)]
		_, okPending := inPending[favoriteKey(sourceTypeSehuatang, row.RowID)]
		row.IsFavorited = okDownload
		row.IsInDownload = okDownload
		row.IsInPending = okPending
	}
	return nil
}

func favoriteKey(sourceType string, sourceRowID int64) string {
	return strings.TrimSpace(sourceType) + ":" + fmt.Sprintf("%d", sourceRowID)
}

func normalizeInfoHash(hash string) string {
	return strings.ToUpper(strings.TrimSpace(hash))
}
