package fetchjavbus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/fetchsite"
)

const (
	fetchStatusRunning = fetchsite.FetchStatusRunning
	fetchStatusSuccess = fetchsite.FetchStatusSuccess
	fetchStatusFailed  = fetchsite.FetchStatusFailed
)

func (s *Service) FetchMovieMagnets(ctx context.Context, movieJavID, movieName string) ([]*moviex.TJavbusMagnet, error) {
	detailURL, detailPayload, detailAttempts, err := s.fetchDetailPage(ctx, movieName)
	if err != nil {
		_ = s.saveFetchFailure(ctx, movieJavID, movieName, detailURL, detailAttempts, err)
		return nil, err
	}
	if err := s.siteSvc.SleepRequest(ctx, fetchsite.FetchSiteCodeJavbus); err != nil {
		return nil, err
	}

	ajaxURL, err := detailPayload.buildAjaxURL(s.siteSvc)
	if err != nil {
		_ = s.saveFetchFailure(ctx, movieJavID, movieName, detailURL, 1, err)
		return nil, err
	}

	resp, attempts, err := s.siteSvc.FetchWithRetry(ctx, fetchsite.FetchSiteCodeJavbus, ajaxURL, fetchsite.RequestOptions{
		Referer: detailURL,
		Headers: map[string]string{
			"X-Requested-With": "XMLHttpRequest",
		},
	}, nil)
	if err != nil {
		_ = s.saveFetchFailure(ctx, movieJavID, movieName, ajaxURL, attempts, err)
		return nil, err
	}
	if err := fetchsite.RequireHTTP200(resp.Status); err != nil {
		_ = s.saveFetchFailure(ctx, movieJavID, movieName, ajaxURL, attempts, err)
		return nil, err
	}

	magnets, err := parseMagnetRows(movieJavID, string(resp.Body))
	if err != nil {
		_ = s.saveFetchFailure(ctx, movieJavID, movieName, ajaxURL, attempts, err)
		return nil, err
	}
	if err := s.saveMagnets(ctx, movieJavID, movieName, ajaxURL, attempts, magnets); err != nil {
		return nil, err
	}
	return magnets, nil
}

func (s *Service) fetchDetailPage(ctx context.Context, movieName string) (string, *detailPagePayload, int64, error) {
	normalizedCode := strings.ToLower(fetchsite.NormalizeMovieName(movieName))
	detailURL, err := s.siteSvc.BuildURL(fetchsite.FetchSiteCodeJavbus, normalizedCode)
	if err != nil {
		return "", nil, 0, err
	}

	resp, attempts, err := s.siteSvc.FetchWithRetry(ctx, fetchsite.FetchSiteCodeJavbus, detailURL, fetchsite.RequestOptions{}, func(resp *fetchsite.Response) error {
		switch resp.Status {
		case 200, 302:
		default:
			return fmt.Errorf("unexpected javbus detail status: %d", resp.Status)
		}
		if len(resp.Body) == 0 {
			return fmt.Errorf("empty javbus detail body")
		}
		return nil
	})
	if err != nil {
		return detailURL, nil, attempts, err
	}

	payload, err := parseDetailPagePayload(string(resp.Body))
	if err != nil {
		return detailURL, nil, attempts, err
	}
	return detailURL, payload, attempts, nil
}

func (s *Service) saveMagnets(ctx context.Context, movieJavID, movieName, sourceURL string, tryCount int64, magnets []*moviex.TJavbusMagnet) error {
	now := time.Now().Unix()
	fetchRow, err := s.ensureFetchRow(ctx, movieJavID, movieName, sourceURL, now)
	if err != nil {
		return err
	}

	for _, magnet := range magnets {
		oldRow, findErr := s.deps.JavbusMagnetModel.FindOneByMovieJavIdInfoHash(ctx, magnet.MovieJavId, magnet.InfoHash)
		if findErr != nil && findErr != moviex.ErrNotFound {
			return findErr
		}
		if oldRow == nil {
			magnet.CreatedOn = now
			magnet.UpdatedOn = now
			magnet.LastSeenTime = now
			if _, insertErr := s.deps.JavbusMagnetModel.Insert(ctx, magnet); insertErr != nil {
				return insertErr
			}
			continue
		}

		oldRow.MagnetName = magnet.MagnetName
		oldRow.SizeBytes = magnet.SizeBytes
		oldRow.ShareDate = magnet.ShareDate
		oldRow.HasHd = magnet.HasHd
		oldRow.HasSubtitle = magnet.HasSubtitle
		oldRow.RowSort = magnet.RowSort
		oldRow.LastSeenTime = now
		oldRow.UpdatedOn = now
		if updateErr := s.deps.JavbusMagnetModel.Update(ctx, oldRow); updateErr != nil {
			return updateErr
		}
	}

	torrentHashCount, latestPublishTime, err := s.deps.JavbusMagnetModel.SummarizeByMovieJavId(ctx, movieJavID)
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
	fetchRow.LastResultCount = int64(len(magnets))
	fetchRow.TorrentHashCount = torrentHashCount
	fetchRow.LatestPublishTime = latestPublishTime
	fetchRow.SourceUrl = sourceURL
	fetchRow.UpdatedOn = now
	return s.deps.JavbusMagnetFetchModel.Update(ctx, fetchRow)
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
	return s.deps.JavbusMagnetFetchModel.Update(ctx, fetchRow)
}

func (s *Service) ensureFetchRow(ctx context.Context, movieJavID, movieName, sourceURL string, now int64) (*moviex.TJavbusMagnetFetch, error) {
	row, err := s.deps.JavbusMagnetFetchModel.FindOneByMovieJavId(ctx, movieJavID)
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

	row = &moviex.TJavbusMagnetFetch{
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
	result, insertErr := s.deps.JavbusMagnetFetchModel.Insert(ctx, row)
	if insertErr != nil {
		return nil, insertErr
	}
	insertID, idErr := result.LastInsertId()
	if idErr != nil {
		return nil, idErr
	}
	return s.deps.JavbusMagnetFetchModel.FindOne(ctx, insertID)
}
