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

func (s *Service) FetchMovieMagnets(ctx context.Context, movieJavID, movieCode string) ([]*moviex.TJavbusMagnet, error) {
	reportInfoLog(ctx, fmt.Sprintf("JavBus 详情页请求: %s", movieCode))
	detailURL, detailPayload, detailAttempts, err := s.fetchDetailPage(ctx, movieCode)
	if err != nil {
		reportErrorLog(ctx, fmt.Sprintf("JavBus 详情页失败: %s | attempts=%d | err=%v", movieCode, detailAttempts, err))
		_ = s.saveFetchFailure(ctx, movieJavID, movieCode, detailURL, detailAttempts, err)
		return nil, err
	}
	reportInfoLog(ctx, fmt.Sprintf("JavBus 详情页成功: %s | attempts=%d", movieCode, detailAttempts))

	reportInfoLog(ctx, fmt.Sprintf("JavBus Ajax 前等待: %s", movieCode))
	if err := s.siteSvc.SleepRequest(ctx, fetchsite.FetchSiteCodeJavbus); err != nil {
		return nil, err
	}

	ajaxURL, err := detailPayload.buildAjaxURL(s.siteSvc)
	if err != nil {
		reportErrorLog(ctx, fmt.Sprintf("JavBus Ajax URL 构造失败: %s | err=%v", movieCode, err))
		_ = s.saveFetchFailure(ctx, movieJavID, movieCode, detailURL, 1, err)
		return nil, err
	}
	reportInfoLog(ctx, fmt.Sprintf("JavBus Ajax 请求: %s", ajaxURL))

	resp, attempts, err := s.siteSvc.FetchWithRetry(ctx, fetchsite.FetchSiteCodeJavbus, ajaxURL, fetchsite.RequestOptions{
		Referer: detailURL,
		Headers: map[string]string{
			"X-Requested-With": "XMLHttpRequest",
		},
	}, nil)
	if err != nil {
		reportErrorLog(ctx, fmt.Sprintf("JavBus Ajax 失败: %s | attempts=%d | err=%v", movieCode, attempts, err))
		_ = s.saveFetchFailure(ctx, movieJavID, movieCode, ajaxURL, attempts, err)
		return nil, err
	}
	if err := fetchsite.RequireHTTP200(resp.Status); err != nil {
		reportErrorLog(ctx, fmt.Sprintf("JavBus Ajax 非 200: %s | status=%d | err=%v", movieCode, resp.Status, err))
		_ = s.saveFetchFailure(ctx, movieJavID, movieCode, ajaxURL, attempts, err)
		return nil, err
	}
	reportInfoLog(ctx, fmt.Sprintf("JavBus Ajax 成功: %s | attempts=%d", movieCode, attempts))

	magnets, err := parseMagnetRows(movieJavID, ajaxURL, string(resp.Body))
	if err != nil {
		reportErrorLog(ctx, fmt.Sprintf("JavBus 解析失败: %s | err=%v", movieCode, err))
		_ = s.saveFetchFailure(ctx, movieJavID, movieCode, ajaxURL, attempts, err)
		return nil, err
	}
	reportInfoLog(ctx, fmt.Sprintf("JavBus 解析完成: %s | count=%d", movieCode, len(magnets)))
	if err := s.saveMagnets(ctx, movieJavID, movieCode, ajaxURL, attempts, magnets); err != nil {
		reportErrorLog(ctx, fmt.Sprintf("JavBus 落库失败: %s | err=%v", movieCode, err))
		return nil, err
	}
	reportInfoLog(ctx, fmt.Sprintf("JavBus 落库完成: %s | count=%d", movieCode, len(magnets)))
	return magnets, nil
}

func (s *Service) fetchDetailPage(ctx context.Context, movieCode string) (string, *detailPagePayload, int64, error) {
	normalizedCode := strings.ToLower(fetchsite.NormalizeMovieCode(movieCode))
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

func (s *Service) saveMagnets(ctx context.Context, movieJavID, movieCode, sourceURL string, tryCount int64, magnets []*moviex.TJavbusMagnet) error {
	now := time.Now().Unix()
	fetchRow, err := s.ensureFetchRow(ctx, movieJavID, movieCode, sourceURL, now)
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

		oldRow.PageUrl = magnet.PageUrl
		oldRow.MagnetName = magnet.MagnetName
		oldRow.MagnetUrl = magnet.MagnetUrl
		oldRow.Dn = magnet.Dn
		oldRow.SizeBytes = magnet.SizeBytes
		oldRow.SizeText = magnet.SizeText
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

func (s *Service) saveFetchFailure(ctx context.Context, movieJavID, movieCode, sourceURL string, tryCount int64, fetchErr error) error {
	now := time.Now().Unix()
	fetchRow, err := s.ensureFetchRow(ctx, movieJavID, movieCode, sourceURL, now)
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

func (s *Service) ensureFetchRow(ctx context.Context, movieJavID, movieCode, sourceURL string, now int64) (*moviex.TJavbusMagnetFetch, error) {
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
		MovieCode:         movieCode,
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
