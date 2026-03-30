package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

const (
	ingestPrecheckPlanVersion  = 2
	ingestPrecheckPlanFileName = "media_ingest_precheck_plan.json"
)

var ErrIngestPrecheckPlanNotFound = errors.New("media ingest precheck plan not found")

type ingestPrecheckPlan struct {
	Version     int                        `json:"version"`
	RootDir     string                     `json:"root_dir"`
	GeneratedAt int64                      `json:"generated_at"`
	Total       int                        `json:"total"`
	Passed      int                        `json:"passed"`
	Failed      int                        `json:"failed"`
	Entries     []*ingestPrecheckPlanEntry `json:"entries"`
	Checks      []*ingestPrecheckPlanCheck `json:"checks"`
}

type ingestPrecheckPlanEntry struct {
	SourcePath          string `json:"source_path"`
	MovieName           string `json:"movie_name"`
	Ext                 string `json:"ext"`
	HasSub              int64  `json:"has_sub"`
	SelfMake            int64  `json:"self_make"`
	HasMask             int64  `json:"has_mask"`
	MovieJavID          string `json:"movie_jav_id"`
	ReleasingDay        int64  `json:"releasing_day"`
	FavoriteAlbumID     int64  `json:"favorite_album_id"`
	FavoriteSourceType  string `json:"favorite_source_type"`
	FavoriteSourceRowID int64  `json:"favorite_source_row_id"`
	FavoriteInfoHash    string `json:"favorite_info_hash"`
}

type ingestPrecheckPlanCheck struct {
	Status            string `json:"status"`
	RootDir           string `json:"root_dir"`
	SourcePath        string `json:"source_path"`
	MovieName         string `json:"movie_name"`
	TargetFileName    string `json:"target_file_name"`
	TargetDir         string `json:"target_dir"`
	Alias             string `json:"alias"`
	SourceTorrentHash string `json:"source_torrent_hash"`
	Size              int64  `json:"size"`
	BirthTime         int64  `json:"birth_time"`
	TargetPath        string `json:"target_path"`
	FailedPath        string `json:"failed_path"`
	Error             string `json:"error"`
}

func buildIngestPrecheckPlan(layout rootLayout, passPrepared []*ingestPreparedItem, items []*IngestFileItem) *ingestPrecheckPlan {
	plan := &ingestPrecheckPlan{
		Version:     ingestPrecheckPlanVersion,
		RootDir:     layout.rootDir,
		GeneratedAt: time.Now().Unix(),
		Entries:     make([]*ingestPrecheckPlanEntry, 0, len(passPrepared)),
		Checks:      make([]*ingestPrecheckPlanCheck, 0, len(items)),
	}

	for _, prepared := range passPrepared {
		if prepared == nil || prepared.favoriteSource == nil || prepared.favoriteSource.item == nil {
			continue
		}
		sourceItem := prepared.favoriteSource.item
		plan.Entries = append(plan.Entries, &ingestPrecheckPlanEntry{
			SourcePath:          prepared.sourcePath,
			MovieName:           prepared.meta.movieName,
			Ext:                 prepared.meta.ext,
			HasSub:              prepared.meta.hasSub,
			SelfMake:            prepared.meta.selfMake,
			HasMask:             prepared.meta.hasMask,
			MovieJavID:          prepared.movieInfo.javID,
			ReleasingDay:        prepared.movieInfo.releasingDay,
			FavoriteAlbumID:     prepared.favoriteSource.favoriteAlbumID,
			FavoriteSourceType:  sourceItem.SourceType,
			FavoriteSourceRowID: sourceItem.SourceRowId,
			FavoriteInfoHash:    prepared.favoriteSource.infoHash,
		})
	}

	for _, item := range items {
		if item == nil {
			continue
		}
		status := strings.TrimSpace(item.Status)
		if status == "" {
			if strings.TrimSpace(item.Error) == "" {
				status = ingestItemStatusPass
			} else {
				status = ingestItemStatusFail
			}
		}
		if status == ingestItemStatusPass {
			plan.Passed++
		} else {
			plan.Failed++
		}
		plan.Total++

		plan.Checks = append(plan.Checks, &ingestPrecheckPlanCheck{
			Status:            status,
			RootDir:           strings.TrimSpace(item.RootDir),
			SourcePath:        strings.TrimSpace(item.SourcePath),
			MovieName:         strings.TrimSpace(item.MovieName),
			TargetFileName:    strings.TrimSpace(item.TargetFileName),
			TargetDir:         strings.TrimSpace(item.TargetDir),
			Alias:             strings.TrimSpace(item.Alias),
			SourceTorrentHash: strings.TrimSpace(item.SourceTorrentHash),
			Size:              item.Size,
			BirthTime:         item.BirthTime,
			TargetPath:        strings.TrimSpace(item.TargetPath),
			FailedPath:        strings.TrimSpace(item.FailedPath),
			Error:             strings.TrimSpace(item.Error),
		})
	}
	return plan
}

func (s *Service) saveIngestPrecheckPlan(layout rootLayout, plan *ingestPrecheckPlan) error {
	if plan == nil {
		plan = &ingestPrecheckPlan{
			Version:     ingestPrecheckPlanVersion,
			RootDir:     layout.rootDir,
			GeneratedAt: time.Now().Unix(),
			Entries:     []*ingestPrecheckPlanEntry{},
			Checks:      []*ingestPrecheckPlanCheck{},
		}
	}
	if plan.Version <= 0 {
		plan.Version = ingestPrecheckPlanVersion
	}
	if strings.TrimSpace(plan.RootDir) == "" {
		plan.RootDir = layout.rootDir
	}
	if plan.GeneratedAt <= 0 {
		plan.GeneratedAt = time.Now().Unix()
	}
	if plan.Entries == nil {
		plan.Entries = []*ingestPrecheckPlanEntry{}
	}
	if plan.Checks == nil {
		plan.Checks = []*ingestPrecheckPlanCheck{}
	}

	payload, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}

	path := ingestPrecheckPlanPath(layout)
	tmpPath := path + ".tmp"
	if err = os.MkdirAll(filepath.Dir(path), defaultFilePerm); err != nil {
		return err
	}
	if err = os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return err
	}
	if err = os.Rename(tmpPath, path); err != nil {
		return err
	}
	return nil
}

func (s *Service) loadIngestPrecheckPlan(layout rootLayout) (*ingestPrecheckPlan, error) {
	path := ingestPrecheckPlanPath(layout)
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrIngestPrecheckPlanNotFound, path)
		}
		return nil, err
	}

	plan := &ingestPrecheckPlan{}
	if err = json.Unmarshal(payload, plan); err != nil {
		return nil, fmt.Errorf("预处理计划解析失败: %w", err)
	}
	if plan.Version != ingestPrecheckPlanVersion {
		return nil, fmt.Errorf("预处理计划版本不支持: %d", plan.Version)
	}
	if strings.TrimSpace(plan.RootDir) == "" {
		plan.RootDir = layout.rootDir
	}
	if plan.Entries == nil {
		plan.Entries = []*ingestPrecheckPlanEntry{}
	}
	if plan.Checks == nil {
		plan.Checks = []*ingestPrecheckPlanCheck{}
	}
	return plan, nil
}

func (s *Service) clearIngestPrecheckPlan(layout rootLayout) error {
	path := ingestPrecheckPlanPath(layout)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func ingestPrecheckPlanPath(layout rootLayout) string {
	return filepath.Join(layout.tmp, ingestPrecheckPlanFileName)
}

func buildPreparedFromPlanEntry(entry *ingestPrecheckPlanEntry) (*ingestPreparedItem, error) {
	if entry == nil {
		return nil, fmt.Errorf("empty precheck plan entry")
	}
	sourcePath := strings.TrimSpace(entry.SourcePath)
	if sourcePath == "" {
		return nil, fmt.Errorf("empty source_path in precheck plan")
	}
	if _, err := os.Stat(sourcePath); err != nil {
		return nil, fmt.Errorf("source file missing: %s", sourcePath)
	}

	movieName := strings.ToUpper(strings.TrimSpace(entry.MovieName))
	if movieName == "" {
		return nil, fmt.Errorf("empty movie_name in precheck plan: %s", sourcePath)
	}
	movieJavID := strings.TrimSpace(entry.MovieJavID)
	if movieJavID == "" {
		return nil, fmt.Errorf("empty movie_jav_id in precheck plan: %s", sourcePath)
	}
	infoHash := strings.TrimSpace(entry.FavoriteInfoHash)
	if infoHash == "" {
		return nil, fmt.Errorf("empty favorite info_hash in precheck plan: %s", sourcePath)
	}
	sourceType := strings.TrimSpace(entry.FavoriteSourceType)
	if sourceType == "" {
		return nil, fmt.Errorf("empty favorite source_type in precheck plan: %s", sourcePath)
	}
	if entry.FavoriteSourceRowID <= 0 {
		return nil, fmt.Errorf("invalid favorite source_row_id in precheck plan: %s", sourcePath)
	}
	if entry.FavoriteAlbumID <= 0 {
		return nil, fmt.Errorf("invalid favorite album_id in precheck plan: %s", sourcePath)
	}

	meta := rawMovieMeta{
		movieName: movieName,
		ext:       normalizePlanExt(entry.Ext),
		hasSub:    entry.HasSub,
		selfMake:  entry.SelfMake,
		hasMask:   entry.HasMask,
	}
	return &ingestPreparedItem{
		sourcePath: sourcePath,
		meta:       meta,
		movieInfo: movieInfo{
			javID:        movieJavID,
			releasingDay: entry.ReleasingDay,
		},
		favoriteSource: &favoriteAlbumSourceInfo{
			favoriteAlbumID: entry.FavoriteAlbumID,
			infoHash:        infoHash,
			item: &moviex.TmAlbumItem{
				SourceType:  sourceType,
				SourceRowId: entry.FavoriteSourceRowID,
				MovieJavId:  movieJavID,
				InfoHash:    infoHash,
				MovieName:   movieName,
			},
		},
	}, nil
}

func normalizePlanExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ".mp4"
	}
	if strings.HasPrefix(ext, ".") {
		return ext
	}
	return "." + ext
}

func safePlanSourcePath(entry *ingestPrecheckPlanEntry) string {
	if entry == nil {
		return ""
	}
	return strings.TrimSpace(entry.SourcePath)
}

func safePlanMovieName(entry *ingestPrecheckPlanEntry) string {
	if entry == nil {
		return ""
	}
	return strings.TrimSpace(entry.MovieName)
}
