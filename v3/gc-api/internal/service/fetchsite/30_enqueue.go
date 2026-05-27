package fetchsite

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) EnsureFetchTasksForMovie(ctx context.Context, movieJavID, movieName string, releaseDate int64) error {
	now := time.Now().Unix()

	if err := s.ensureJavbusFetchTask(ctx, movieJavID, movieName, releaseDate, now); err != nil {
		return err
	}
	if err := s.ensureSukebeiFetchTask(ctx, movieJavID, movieName, releaseDate, now); err != nil {
		return err
	}
	return nil
}

func (s *Service) ensureJavbusFetchTask(ctx context.Context, movieJavID, movieName string, releaseDate int64, now int64) error {
	if row, err := s.deps.JavbusMagnetFetchModel.FindOneByMovieJavId(ctx, movieJavID); err == nil {
		if row.ReleaseDate != releaseDate {
			row.ReleaseDate = releaseDate
			row.UpdatedOn = now
			return s.deps.JavbusMagnetFetchModel.Update(ctx, row)
		}
		return nil
	} else if err != moviex.ErrNotFound {
		return err
	}

	sourceURL, err := s.BuildURL(FetchSiteCodeJavbus, strings.ToLower(NormalizeMovieName(movieName)))
	if err != nil {
		return err
	}

	row := &moviex.TJavbusMagnetFetch{
		MovieJavId:      movieJavID,
		MovieName:       movieName,
		ReleaseDate:     releaseDate,
		FetchStatus:     FetchStatusPending,
		TryCount:        0,
		LastFetchTime:   0,
		LastSuccessTime: 0,
		LastError:       "",
		LastResultCount: 0,
		SourceUrl:       sourceURL,
		CreatedOn:       now,
		UpdatedOn:       now,
	}
	_, err = s.deps.JavbusMagnetFetchModel.Insert(ctx, row)
	return err
}

func (s *Service) ensureSukebeiFetchTask(ctx context.Context, movieJavID, movieName string, releaseDate int64, now int64) error {
	if row, err := s.deps.SukebeiTorrentFetchModel.FindOneByMovieJavId(ctx, movieJavID); err == nil {
		if row.ReleaseDate != releaseDate {
			row.ReleaseDate = releaseDate
			row.UpdatedOn = now
			return s.deps.SukebeiTorrentFetchModel.Update(ctx, row)
		}
		return nil
	} else if err != moviex.ErrNotFound {
		return err
	}

	queryText := BuildSukebeiQuery(movieName)
	if strings.TrimSpace(queryText) == "" {
		return fmt.Errorf("invalid movie code for sukebei query: %s", movieName)
	}

	baseURL, err := s.BuildURL(FetchSiteCodeSukebei)
	if err != nil {
		return err
	}
	values := url.Values{}
	values.Set("f", "0")
	values.Set("c", "0_0")
	values.Set("q", queryText)

	row := &moviex.TSukebeiTorrentFetch{
		MovieJavId:      movieJavID,
		MovieName:       movieName,
		ReleaseDate:     releaseDate,
		FetchStatus:     FetchStatusPending,
		TryCount:        0,
		LastFetchTime:   0,
		LastSuccessTime: 0,
		LastError:       "",
		LastResultCount: 0,
		SourceUrl:       baseURL + "/?" + values.Encode(),
		CreatedOn:       now,
		UpdatedOn:       now,
	}
	_, err = s.deps.SukebeiTorrentFetchModel.Insert(ctx, row)
	return err
}
