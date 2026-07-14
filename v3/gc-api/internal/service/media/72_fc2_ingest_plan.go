package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

var ErrFC2IngestPlanNotFound = errors.New("fc2 media ingest precheck plan not found")

type fc2IngestPlan struct {
	Version     int                   `json:"version"`
	SourceDir   string                `json:"source_dir"`
	RootDir     string                `json:"root_dir"`
	GeneratedAt int64                 `json:"generated_at"`
	Total       int                   `json:"total"`
	Passed      int                   `json:"passed"`
	Failed      int                   `json:"failed"`
	Entries     []*fc2IngestPlanEntry `json:"entries"`
	Checks      []*FC2IngestFileItem  `json:"checks"`
}

type fc2IngestPlanEntry struct {
	SourcePath          string `json:"source_path"`
	FC2Code             string `json:"fc2_code"`
	FC2Key              string `json:"fc2_key"`
	Ext                 string `json:"ext"`
	HasSub              int64  `json:"has_sub"`
	SelfMake            int64  `json:"self_make"`
	HasMask             int64  `json:"has_mask"`
	FC2MovieID          int64  `json:"fc2_movie_id"`
	SourceThreadID      int64  `json:"source_thread_id"`
	SourceTorrentHash   string `json:"source_torrent_hash"`
	FavoriteAlbumID     int64  `json:"favorite_album_id"`
	FavoriteItemID      int64  `json:"favorite_item_id"`
	FavoriteSourceType  string `json:"favorite_source_type"`
	FavoriteSourceRowID int64  `json:"favorite_source_row_id"`
	FavoriteInfoHash    string `json:"favorite_info_hash"`
}

type FC2IngestPlanSnapshot struct {
	Roots       int                  `json:"roots"`
	HasPlan     bool                 `json:"has_plan"`
	GeneratedAt int64                `json:"generated_at"`
	Total       int                  `json:"total"`
	Passed      int                  `json:"passed"`
	Failed      int                  `json:"failed"`
	Items       []*FC2IngestFileItem `json:"items"`
}

func buildFC2IngestPlan(root fc2MediaRoot, prepared []*fc2IngestPreparedItem, checks []*FC2IngestFileItem) *fc2IngestPlan {
	plan := &fc2IngestPlan{
		Version:     fc2IngestPlanVersion,
		SourceDir:   root.sourceDir,
		RootDir:     root.rootDir,
		GeneratedAt: time.Now().Unix(),
		Entries:     make([]*fc2IngestPlanEntry, 0, len(prepared)),
		Checks:      checks,
	}
	for _, item := range prepared {
		if item == nil || item.favoriteSource == nil || item.favoriteSource.item == nil {
			continue
		}
		sourceItem := item.favoriteSource.item
		plan.Entries = append(plan.Entries, &fc2IngestPlanEntry{
			SourcePath:          item.sourcePath,
			FC2Code:             item.meta.code,
			FC2Key:              item.meta.key,
			Ext:                 item.meta.ext,
			HasSub:              item.meta.hasSub,
			SelfMake:            item.meta.selfMake,
			HasMask:             item.hasMask,
			FC2MovieID:          item.fc2MovieID,
			SourceThreadID:      item.sourceThreadID,
			SourceTorrentHash:   item.sourceTorrentHash,
			FavoriteAlbumID:     item.favoriteSource.favoriteAlbumID,
			FavoriteItemID:      sourceItem.Id,
			FavoriteSourceType:  sourceItem.SourceType,
			FavoriteSourceRowID: sourceItem.SourceRowId,
			FavoriteInfoHash:    item.favoriteSource.infoHash,
		})
	}
	for _, item := range checks {
		if item == nil {
			continue
		}
		plan.Total++
		if strings.TrimSpace(item.Error) == "" {
			plan.Passed++
		} else {
			plan.Failed++
		}
	}
	return plan
}

func saveFC2IngestPlan(layout rootLayout, plan *fc2IngestPlan) error {
	if plan == nil {
		plan = &fc2IngestPlan{
			Version:     fc2IngestPlanVersion,
			SourceDir:   layout.ingestNew,
			RootDir:     layout.rootDir,
			GeneratedAt: time.Now().Unix(),
			Entries:     []*fc2IngestPlanEntry{},
			Checks:      []*FC2IngestFileItem{},
		}
	}
	if plan.Version <= 0 {
		plan.Version = fc2IngestPlanVersion
	}
	if strings.TrimSpace(plan.SourceDir) == "" {
		plan.SourceDir = layout.ingestNew
	}
	if strings.TrimSpace(plan.RootDir) == "" {
		plan.RootDir = layout.rootDir
	}
	if plan.GeneratedAt <= 0 {
		plan.GeneratedAt = time.Now().Unix()
	}
	if plan.Entries == nil {
		plan.Entries = []*fc2IngestPlanEntry{}
	}
	if plan.Checks == nil {
		plan.Checks = []*FC2IngestFileItem{}
	}
	payload, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	path := fc2IngestPlanPath(layout)
	tmpPath := path + ".tmp"
	if err := os.MkdirAll(filepath.Dir(path), defaultFilePerm); err != nil {
		return err
	}
	if err := os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func loadFC2IngestPlan(layout rootLayout) (*fc2IngestPlan, error) {
	path := fc2IngestPlanPath(layout)
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrFC2IngestPlanNotFound, path)
		}
		return nil, err
	}
	plan := &fc2IngestPlan{}
	if err := json.Unmarshal(payload, plan); err != nil {
		return nil, fmt.Errorf("FC2预处理计划解析失败: %w", err)
	}
	if plan.Version != fc2IngestPlanVersion {
		return nil, fmt.Errorf("FC2预处理计划版本不支持: %d", plan.Version)
	}
	if strings.TrimSpace(plan.SourceDir) == "" {
		plan.SourceDir = layout.ingestNew
	}
	if strings.TrimSpace(plan.RootDir) == "" {
		plan.RootDir = layout.rootDir
	}
	if plan.Entries == nil {
		plan.Entries = []*fc2IngestPlanEntry{}
	}
	if plan.Checks == nil {
		plan.Checks = []*FC2IngestFileItem{}
	}
	return plan, nil
}

func clearFC2IngestPlan(layout rootLayout) error {
	if err := os.Remove(fc2IngestPlanPath(layout)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func fc2IngestPlanPath(layout rootLayout) string {
	return filepath.Join(layout.tmp, fc2IngestPlanFileName)
}

func buildFC2PreparedFromPlan(root fc2MediaRoot, entry *fc2IngestPlanEntry) (*fc2IngestPreparedItem, error) {
	if entry == nil {
		return nil, fmt.Errorf("FC2预处理计划条目为空")
	}
	if filepath.Clean(strings.TrimSpace(entry.SourcePath)) == "." {
		return nil, fmt.Errorf("FC2预处理计划缺少源文件")
	}
	if _, err := os.Stat(entry.SourcePath); err != nil {
		return nil, fmt.Errorf("FC2源文件不存在: %s", entry.SourcePath)
	}
	if strings.TrimSpace(entry.FC2Code) == "" || strings.TrimSpace(entry.FC2Key) == "" || entry.FC2MovieID <= 0 || entry.SourceThreadID <= 0 || strings.TrimSpace(entry.SourceTorrentHash) == "" {
		return nil, fmt.Errorf("FC2预处理计划来源字段不完整: %s", entry.SourcePath)
	}
	if entry.FavoriteAlbumID <= 0 || entry.FavoriteItemID <= 0 || entry.FavoriteSourceRowID <= 0 || strings.TrimSpace(entry.FavoriteSourceType) == "" || strings.TrimSpace(entry.FavoriteInfoHash) == "" {
		return nil, fmt.Errorf("FC2预处理计划相册来源字段不完整: %s", entry.SourcePath)
	}
	return &fc2IngestPreparedItem{
		root:       root,
		sourcePath: entry.SourcePath,
		meta: fc2RawMediaMeta{
			code:     entry.FC2Code,
			key:      entry.FC2Key,
			ext:      normalizePlanExt(entry.Ext),
			hasSub:   entry.HasSub,
			selfMake: entry.SelfMake,
		},
		fc2MovieID:        entry.FC2MovieID,
		sourceThreadID:    entry.SourceThreadID,
		sourceTorrentHash: strings.ToUpper(strings.TrimSpace(entry.SourceTorrentHash)),
		favoriteSource: &favoriteAlbumSourceInfo{
			favoriteAlbumID: entry.FavoriteAlbumID,
			infoHash:        strings.ToUpper(strings.TrimSpace(entry.FavoriteInfoHash)),
			item: &moviex.TmAlbumItem{
				Id:          entry.FavoriteItemID,
				AlbumId:     entry.FavoriteAlbumID,
				SourceType:  strings.TrimSpace(entry.FavoriteSourceType),
				SourceRowId: entry.FavoriteSourceRowID,
				MovieName:   strings.TrimSpace(entry.FC2Code),
				InfoHash:    strings.ToUpper(strings.TrimSpace(entry.FavoriteInfoHash)),
			},
		},
		hasMask: entry.HasMask,
	}, nil
}

func (s *Service) ReadFC2IngestPlanSnapshot(ctx context.Context) (*FC2IngestPlanSnapshot, error) {
	_ = ctx
	roots := s.fc2MediaRoots()
	snapshot := &FC2IngestPlanSnapshot{Roots: len(roots), Items: make([]*FC2IngestFileItem, 0, 64)}
	for _, root := range roots {
		plan, err := loadFC2IngestPlan(buildRootLayout(root.rootDir))
		if err != nil {
			if errors.Is(err, ErrFC2IngestPlanNotFound) {
				continue
			}
			return nil, err
		}
		snapshot.HasPlan = true
		if plan.GeneratedAt > snapshot.GeneratedAt {
			snapshot.GeneratedAt = plan.GeneratedAt
		}
		snapshot.Total += plan.Total
		snapshot.Passed += plan.Passed
		snapshot.Failed += plan.Failed
		snapshot.Items = append(snapshot.Items, plan.Checks...)
	}
	sort.SliceStable(snapshot.Items, func(i, j int) bool {
		leftFail := strings.TrimSpace(snapshot.Items[i].Error) != ""
		rightFail := strings.TrimSpace(snapshot.Items[j].Error) != ""
		if leftFail != rightFail {
			return leftFail
		}
		leftCode := strings.TrimSpace(snapshot.Items[i].FC2Code)
		rightCode := strings.TrimSpace(snapshot.Items[j].FC2Code)
		if leftCode != rightCode {
			return leftCode < rightCode
		}
		return strings.TrimSpace(snapshot.Items[i].SourcePath) < strings.TrimSpace(snapshot.Items[j].SourcePath)
	})
	return snapshot, nil
}

func (s *Service) ClearFC2IngestPlan(ctx context.Context) error {
	_ = ctx
	for _, root := range s.fc2MediaRoots() {
		if err := clearFC2IngestPlan(buildRootLayout(root.rootDir)); err != nil {
			return err
		}
	}
	return nil
}
