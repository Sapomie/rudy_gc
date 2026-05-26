package media

import (
	"context"
	"errors"
	"sort"
	"strings"
)

type IngestPlanSnapshot struct {
	Roots       int               `json:"roots"`
	HasPlan     bool              `json:"has_plan"`
	GeneratedAt int64             `json:"generated_at"`
	Total       int               `json:"total"`
	Passed      int               `json:"passed"`
	Failed      int               `json:"failed"`
	Items       []*IngestFileItem `json:"items"`
}

func (s *Service) ReadIngestPlanSnapshot(ctx context.Context) (*IngestPlanSnapshot, error) {
	_ = ctx

	roots := s.mediaRoots()
	out := &IngestPlanSnapshot{
		Roots: len(roots),
		Items: make([]*IngestFileItem, 0, 64),
	}
	if len(roots) == 0 {
		return out, nil
	}

	for _, root := range roots {
		layout := buildRootLayout(root)
		plan, err := s.loadIngestPrecheckPlan(layout)
		if err != nil {
			if errors.Is(err, ErrIngestPrecheckPlanNotFound) {
				continue
			}
			return nil, err
		}

		out.HasPlan = true
		if plan.GeneratedAt > out.GeneratedAt {
			out.GeneratedAt = plan.GeneratedAt
		}
		out.Total += plan.Total
		out.Passed += plan.Passed
		out.Failed += plan.Failed

		for _, check := range plan.Checks {
			if check == nil {
				continue
			}
			status := strings.TrimSpace(check.Status)
			if status == "" {
				if strings.TrimSpace(check.Error) == "" {
					status = ingestItemStatusPass
				} else {
					status = ingestItemStatusFail
				}
			}
			out.Items = append(out.Items, &IngestFileItem{
				Status:            status,
				RootDir:           check.RootDir,
				SourcePath:        check.SourcePath,
				MovieName:         check.MovieName,
				TargetFileName:    check.TargetFileName,
				TargetDir:         check.TargetDir,
				Alias:             check.Alias,
				SourceTorrentHash: check.SourceTorrentHash,
				Size:              check.Size,
				BirthTime:         check.BirthTime,
				TargetPath:        check.TargetPath,
				FailedPath:        check.FailedPath,
				Error:             check.Error,
			})
		}
	}

	sort.SliceStable(out.Items, func(i, j int) bool {
		leftFail := strings.TrimSpace(out.Items[i].Error) != ""
		rightFail := strings.TrimSpace(out.Items[j].Error) != ""
		if leftFail != rightFail {
			return leftFail
		}
		leftMovie := strings.TrimSpace(out.Items[i].MovieName)
		rightMovie := strings.TrimSpace(out.Items[j].MovieName)
		if leftMovie != rightMovie {
			return leftMovie < rightMovie
		}
		return strings.TrimSpace(out.Items[i].SourcePath) < strings.TrimSpace(out.Items[j].SourcePath)
	})

	return out, nil
}

func (s *Service) ClearIngestPlanSnapshot(ctx context.Context) error {
	_ = ctx

	roots := s.mediaRoots()
	for _, root := range roots {
		layout := buildRootLayout(root)
		if err := s.clearIngestPrecheckPlan(layout); err != nil {
			return err
		}
	}
	return nil
}
