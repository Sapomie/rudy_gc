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
)

const scMediaMovePlanVersion = 1

var ErrScMediaMovePlanNotFound = errors.New("sc media move plan not found")

type ScMediaMovePlanSnapshot struct {
	Mode        string             `json:"mode"`
	ScName      string             `json:"sc_name"`
	HasPlan     bool               `json:"has_plan"`
	GeneratedAt int64              `json:"generated_at"`
	Total       int                `json:"total"`
	Movable     int                `json:"movable"`
	Skipped     int                `json:"skipped"`
	Failed      int                `json:"failed"`
	Items       []*ScMediaMoveItem `json:"items"`
}

type scMediaMovePlan struct {
	Version     int                     `json:"version"`
	Mode        string                  `json:"mode"`
	ScName      string                  `json:"sc_name"`
	GeneratedAt int64                   `json:"generated_at"`
	Total       int                     `json:"total"`
	Movable     int                     `json:"movable"`
	Skipped     int                     `json:"skipped"`
	Failed      int                     `json:"failed"`
	Entries     []*scMediaMovePlanEntry `json:"entries"`
	Checks      []*scMediaMovePlanCheck `json:"checks"`
}

type scMediaMovePlanEntry struct {
	MovieJavId string   `json:"movie_jav_id"`
	MovieName  string   `json:"movie_name"`
	ScNames    []string `json:"sc_names"`
	RootDir    string   `json:"root_dir"`
	SourcePath string   `json:"source_path"`
}

type scMediaMovePlanCheck struct {
	Status     string   `json:"status"`
	MovieJavId string   `json:"movie_jav_id"`
	MovieName  string   `json:"movie_name"`
	ScNames    []string `json:"sc_names"`
	RootDir    string   `json:"root_dir"`
	SourcePath string   `json:"source_path"`
	TargetDir  string   `json:"target_dir"`
	TargetPath string   `json:"target_path"`
	Error      string   `json:"error"`
}

func (s *Service) ReadScMediaMovePlanSnapshot(ctx context.Context, mode, scName string) (*ScMediaMovePlanSnapshot, error) {
	_ = ctx

	mode = normalizeScMediaMoveMode(mode)
	scName = strings.TrimSpace(scName)
	out := &ScMediaMovePlanSnapshot{
		Mode:   mode,
		ScName: scName,
		Items:  []*ScMediaMoveItem{},
	}
	if mode == scMediaMoveModeSingle && scName == "" {
		return out, nil
	}

	plan, err := s.loadScMediaMovePlan(mode, scName)
	if err != nil {
		if errors.Is(err, ErrScMediaMovePlanNotFound) {
			return out, nil
		}
		return nil, err
	}

	out.HasPlan = true
	out.Mode = normalizeScMediaMoveMode(plan.Mode)
	out.GeneratedAt = plan.GeneratedAt
	out.Total = plan.Total
	out.Movable = plan.Movable
	out.Skipped = plan.Skipped
	out.Failed = plan.Failed
	for _, check := range plan.Checks {
		if check == nil {
			continue
		}
		out.Items = append(out.Items, &ScMediaMoveItem{
			Status:     strings.TrimSpace(check.Status),
			MovieJavId: strings.TrimSpace(check.MovieJavId),
			MovieName:  strings.TrimSpace(check.MovieName),
			ScNames:    normalizeScMediaMoveItemScNames(out.Mode, out.ScName, check.ScNames),
			RootDir:    strings.TrimSpace(check.RootDir),
			SourcePath: strings.TrimSpace(check.SourcePath),
			TargetDir:  strings.TrimSpace(check.TargetDir),
			TargetPath: strings.TrimSpace(check.TargetPath),
			Error:      strings.TrimSpace(check.Error),
		})
	}

	sort.SliceStable(out.Items, func(i, j int) bool {
		left := out.Items[i]
		right := out.Items[j]
		if rankScMediaMoveStatus(left.Status) != rankScMediaMoveStatus(right.Status) {
			return rankScMediaMoveStatus(left.Status) < rankScMediaMoveStatus(right.Status)
		}
		if left.MovieName != right.MovieName {
			return left.MovieName < right.MovieName
		}
		return left.SourcePath < right.SourcePath
	})

	return out, nil
}

func (s *Service) ClearScMediaMovePlanSnapshot(ctx context.Context, mode, scName string) error {
	_ = ctx

	mode = normalizeScMediaMoveMode(mode)
	scName = strings.TrimSpace(scName)
	if mode == scMediaMoveModeSingle && scName == "" {
		return nil
	}
	return s.clearScMediaMovePlan(mode, scName)
}

func buildScMediaMovePlan(mode, scName string, generatedAt int64, items []*ScMediaMoveItem) *scMediaMovePlan {
	plan := &scMediaMovePlan{
		Version:     scMediaMovePlanVersion,
		Mode:        normalizeScMediaMoveMode(mode),
		ScName:      strings.TrimSpace(scName),
		GeneratedAt: generatedAt,
		Entries:     []*scMediaMovePlanEntry{},
		Checks:      []*scMediaMovePlanCheck{},
	}
	if plan.GeneratedAt <= 0 {
		plan.GeneratedAt = time.Now().Unix()
	}

	for _, item := range items {
		if item == nil {
			continue
		}
		status := normalizeScMediaMoveStatus(item.Status)
		switch status {
		case scMediaMoveStatusPass:
			plan.Movable++
			plan.Entries = append(plan.Entries, &scMediaMovePlanEntry{
				MovieJavId: strings.TrimSpace(item.MovieJavId),
				MovieName:  strings.TrimSpace(item.MovieName),
				ScNames:    append([]string{}, item.ScNames...),
				RootDir:    strings.TrimSpace(item.RootDir),
				SourcePath: strings.TrimSpace(item.SourcePath),
			})
		case scMediaMoveStatusSkip:
			plan.Skipped++
		default:
			plan.Failed++
		}
		plan.Total++
		plan.Checks = append(plan.Checks, &scMediaMovePlanCheck{
			Status:     status,
			MovieJavId: strings.TrimSpace(item.MovieJavId),
			MovieName:  strings.TrimSpace(item.MovieName),
			ScNames:    append([]string{}, item.ScNames...),
			RootDir:    strings.TrimSpace(item.RootDir),
			SourcePath: strings.TrimSpace(item.SourcePath),
			TargetDir:  strings.TrimSpace(item.TargetDir),
			TargetPath: strings.TrimSpace(item.TargetPath),
			Error:      strings.TrimSpace(item.Error),
		})
	}
	return plan
}

func (s *Service) saveScMediaMovePlan(mode, scName string, plan *scMediaMovePlan) error {
	layout, err := s.scMediaPlanLayout()
	if err != nil {
		return err
	}
	if err = ensureRootLayout(layout); err != nil {
		return err
	}

	if plan == nil {
		plan = &scMediaMovePlan{
			Version:     scMediaMovePlanVersion,
			Mode:        normalizeScMediaMoveMode(mode),
			ScName:      strings.TrimSpace(scName),
			GeneratedAt: time.Now().Unix(),
			Entries:     []*scMediaMovePlanEntry{},
			Checks:      []*scMediaMovePlanCheck{},
		}
	}
	if plan.Version <= 0 {
		plan.Version = scMediaMovePlanVersion
	}
	plan.Mode = normalizeScMediaMoveMode(plan.Mode)
	if strings.TrimSpace(plan.ScName) == "" {
		plan.ScName = strings.TrimSpace(scName)
	}
	if plan.GeneratedAt <= 0 {
		plan.GeneratedAt = time.Now().Unix()
	}
	if plan.Entries == nil {
		plan.Entries = []*scMediaMovePlanEntry{}
	}
	if plan.Checks == nil {
		plan.Checks = []*scMediaMovePlanCheck{}
	}

	payload, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}

	path := scMediaMovePlanPath(layout, plan.Mode, plan.ScName)
	tmpPath := path + ".tmp"
	if err = os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func (s *Service) loadScMediaMovePlan(mode, scName string) (*scMediaMovePlan, error) {
	layout, err := s.scMediaPlanLayout()
	if err != nil {
		return nil, err
	}

	mode = normalizeScMediaMoveMode(mode)
	payload, err := os.ReadFile(scMediaMovePlanPath(layout, mode, scName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrScMediaMovePlanNotFound, scMediaMoveScopeText(mode, scName))
		}
		return nil, err
	}

	plan := &scMediaMovePlan{}
	if err = json.Unmarshal(payload, plan); err != nil {
		return nil, fmt.Errorf("sc media 计划解析失败: %w", err)
	}
	if plan.Version != scMediaMovePlanVersion {
		return nil, fmt.Errorf("sc media 计划版本不支持: %d", plan.Version)
	}
	plan.Mode = normalizeScMediaMoveMode(plan.Mode)
	if plan.Mode == "" {
		plan.Mode = mode
	}
	if plan.Entries == nil {
		plan.Entries = []*scMediaMovePlanEntry{}
	}
	if plan.Checks == nil {
		plan.Checks = []*scMediaMovePlanCheck{}
	}
	return plan, nil
}

func (s *Service) clearScMediaMovePlan(mode, scName string) error {
	layout, err := s.scMediaPlanLayout()
	if err != nil {
		return err
	}
	path := scMediaMovePlanPath(layout, mode, scName)
	if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *Service) scMediaPlanLayout() (rootLayout, error) {
	roots := s.mediaRoots()
	if len(roots) == 0 {
		return rootLayout{}, fmt.Errorf("media.rootDirs 未配置")
	}
	return buildRootLayout(roots[0]), nil
}

func scMediaMovePlanPath(layout rootLayout, mode, scName string) string {
	mode = normalizeScMediaMoveMode(mode)
	if mode == scMediaMoveModeAll {
		return filepath.Join(layout.tmp, "sc_media_move_plan__all__.json")
	}
	return filepath.Join(layout.tmp, "sc_media_move_plan_"+sanitizeScMediaPlanToken(scName)+".json")
}

func sanitizeScMediaPlanToken(scName string) string {
	scName = strings.TrimSpace(scName)
	if scName == "" {
		return "default"
	}
	var b strings.Builder
	for _, r := range scName {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	token := strings.Trim(b.String(), "_")
	if token == "" {
		return "default"
	}
	return token
}

func normalizeScMediaMoveItemScNames(mode, scName string, scNames []string) []string {
	if len(scNames) == 0 {
		if normalizeScMediaMoveMode(mode) == scMediaMoveModeSingle {
			scName = strings.TrimSpace(scName)
			if scName != "" {
				return []string{scName}
			}
		}
		return []string{}
	}

	out := make([]string, 0, len(scNames))
	seen := make(map[string]struct{}, len(scNames))
	for _, scName := range scNames {
		scName = strings.TrimSpace(scName)
		if scName == "" {
			continue
		}
		if _, ok := seen[scName]; ok {
			continue
		}
		seen[scName] = struct{}{}
		out = append(out, scName)
	}
	sort.Strings(out)
	return out
}
