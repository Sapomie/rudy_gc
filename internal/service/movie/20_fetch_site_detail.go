package movie

import (
	"context"
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
		MovieCode:         row.MovieCode,
		ReleaseDate:       tsToDate(row.ReleaseDate),
		FetchStatus:       row.FetchStatus,
		FetchStatusText:   fetchStatusText(row.FetchStatus),
		TryCount:          row.TryCount,
		LastFetchTime:     tsToDateTime(row.LastFetchTime),
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
		MovieCode:         row.MovieCode,
		ReleaseDate:       tsToDate(row.ReleaseDate),
		FetchStatus:       row.FetchStatus,
		FetchStatusText:   fetchStatusText(row.FetchStatus),
		TryCount:          row.TryCount,
		LastFetchTime:     tsToDateTime(row.LastFetchTime),
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

func normalizeInfoHash(hash string) string {
	return strings.ToUpper(strings.TrimSpace(hash))
}
