package fetchsukebei

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/fetchsite"
)

const (
	fetchStatusRunning = fetchsite.FetchStatusRunning
	fetchStatusSuccess = fetchsite.FetchStatusSuccess
	fetchStatusFailed  = fetchsite.FetchStatusFailed
)

func (s *Service) FetchMovieTorrents(ctx context.Context, movieJavID, movieName string) ([]*moviex.TSukebeiTorrent, error) {
	queryText := fetchsite.BuildSukebeiQuery(movieName)
	if queryText == "" {
		return nil, fmt.Errorf("invalid movie code for sukebei query: %s", movieName)
	}

	searchURL, err := s.buildSearchURL(queryText)
	if err != nil {
		return nil, err
	}

	resp, attempts, err := s.siteSvc.FetchWithRetry(ctx, fetchsite.FetchSiteCodeSukebei, searchURL, fetchsite.RequestOptions{}, func(resp *fetchsite.Response) error {
		if err := fetchsite.RequireHTTP200(resp.Status); err != nil {
			return err
		}
		if len(resp.Body) == 0 {
			return fmt.Errorf("empty sukebei search body")
		}
		return nil
	})
	if err != nil {
		_ = s.saveFetchFailure(ctx, movieJavID, movieName, searchURL, attempts, err)
		return nil, err
	}

	rows, err := parseTorrentRows(movieJavID, queryText, searchURL, string(resp.Body))
	if err != nil {
		_ = s.saveFetchFailure(ctx, movieJavID, movieName, searchURL, attempts, err)
		return nil, err
	}
	if err := s.saveTorrents(ctx, movieJavID, movieName, searchURL, attempts, rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) buildSearchURL(queryText string) (string, error) {
	baseURL, err := s.siteSvc.BuildURL(fetchsite.FetchSiteCodeSukebei)
	if err != nil {
		return "", err
	}
	values := url.Values{}
	values.Set("f", "0")
	values.Set("c", "0_0")
	values.Set("q", queryText)
	return baseURL + "/?" + values.Encode(), nil
}

func (s *Service) saveTorrents(ctx context.Context, movieJavID, movieName, sourceURL string, tryCount int64, rows []*moviex.TSukebeiTorrent) error {
	now := time.Now().Unix()
	fetchRow, err := s.ensureFetchRow(ctx, movieJavID, movieName, sourceURL, now)
	if err != nil {
		return err
	}

	for _, row := range rows {
		oldRow, findErr := s.deps.SukebeiTorrentModel.FindOneByViewId(ctx, row.ViewId)
		if findErr != nil && findErr != moviex.ErrNotFound {
			return findErr
		}
		if oldRow == nil {
			row.CreatedOn = now
			row.UpdatedOn = now
			row.LastSeenTime = now
			if _, insertErr := s.deps.SukebeiTorrentModel.Insert(ctx, row); insertErr != nil {
				return insertErr
			}
			continue
		}

		oldRow.MovieJavId = row.MovieJavId
		oldRow.QueryText = row.QueryText
		oldRow.SearchUrl = row.SearchUrl
		oldRow.TorrentTitle = row.TorrentTitle
		oldRow.ViewUrl = row.ViewUrl
		oldRow.TorrentUrl = row.TorrentUrl
		oldRow.MagnetUrl = row.MagnetUrl
		oldRow.InfoHash = row.InfoHash
		oldRow.Dn = row.Dn
		oldRow.SizeBytes = row.SizeBytes
		oldRow.SizeText = row.SizeText
		oldRow.PublishTime = row.PublishTime
		oldRow.Seeders = row.Seeders
		oldRow.Leechers = row.Leechers
		oldRow.Completed = row.Completed
		oldRow.LastSeenTime = now
		oldRow.UpdatedOn = now
		if updateErr := s.deps.SukebeiTorrentModel.Update(ctx, oldRow); updateErr != nil {
			return updateErr
		}
	}

	torrentHashCount, latestPublishTime, err := s.deps.SukebeiTorrentModel.SummarizeByMovieJavId(ctx, movieJavID)
	if err != nil {
		return err
	}

	fetchRow.FetchStatus = fetchStatusSuccess
	fetchRow.TryCount = fetchRow.TryCount + tryCount
	fetchRow.LastFetchTime = now
	fetchRow.LastSuccessTime = now
	fetchRow.LastError = ""
	if fetchRow.ReleaseDate == 0 {
		releaseDate, err := s.loadMovieReleaseDate(ctx, movieJavID)
		if err != nil {
			return err
		}
		fetchRow.ReleaseDate = releaseDate
	}
	fetchRow.LastResultCount = int64(len(rows))
	fetchRow.TorrentHashCount = torrentHashCount
	fetchRow.LatestPublishTime = latestPublishTime
	fetchRow.SourceUrl = sourceURL
	fetchRow.UpdatedOn = now
	return s.deps.SukebeiTorrentFetchModel.Update(ctx, fetchRow)
}

func (s *Service) saveFetchFailure(ctx context.Context, movieJavID, movieName, sourceURL string, tryCount int64, fetchErr error) error {
	now := time.Now().Unix()
	fetchRow, err := s.ensureFetchRow(ctx, movieJavID, movieName, sourceURL, now)
	if err != nil {
		return err
	}
	fetchRow.FetchStatus = fetchStatusFailed
	fetchRow.TryCount = fetchRow.TryCount + tryCount
	fetchRow.LastFetchTime = now
	fetchRow.LastError = fetchErr.Error()
	if fetchRow.ReleaseDate == 0 {
		releaseDate, err := s.loadMovieReleaseDate(ctx, movieJavID)
		if err != nil {
			return err
		}
		fetchRow.ReleaseDate = releaseDate
	}
	fetchRow.SourceUrl = sourceURL
	fetchRow.UpdatedOn = now
	return s.deps.SukebeiTorrentFetchModel.Update(ctx, fetchRow)
}

func (s *Service) ensureFetchRow(ctx context.Context, movieJavID, movieName, sourceURL string, now int64) (*moviex.TSukebeiTorrentFetch, error) {
	row, err := s.deps.SukebeiTorrentFetchModel.FindOneByMovieJavId(ctx, movieJavID)
	if err == nil {
		return row, nil
	}
	if err != moviex.ErrNotFound {
		return nil, err
	}
	releaseDate, err := s.loadMovieReleaseDate(ctx, movieJavID)
	if err != nil {
		return nil, err
	}

	row = &moviex.TSukebeiTorrentFetch{
		MovieJavId:        movieJavID,
		MovieName:         movieName,
		ReleaseDate:       releaseDate,
		FetchStatus:       fetchStatusRunning,
		TryCount:          0,
		LastFetchTime:     0,
		LastSuccessTime:   0,
		LastError:         "",
		LastResultCount:   0,
		TorrentHashCount:  0,
		LatestPublishTime: 0,
		SourceUrl:         sourceURL,
		CreatedOn:         now,
		UpdatedOn:         now,
	}
	result, insertErr := s.deps.SukebeiTorrentFetchModel.Insert(ctx, row)
	if insertErr != nil {
		return nil, insertErr
	}
	insertID, idErr := result.LastInsertId()
	if idErr != nil {
		return nil, idErr
	}
	return s.deps.SukebeiTorrentFetchModel.FindOne(ctx, insertID)
}
