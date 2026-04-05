package filmmove

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/movie"
)

const (
	defaultPlanTTLSeconds = 6 * 60 * 60
)

type Service struct {
	deps     *dep.Dep
	movieSvc *movie.Service

	mu    sync.Mutex
	seq   uint64
	plans map[string]*movePlan
}

type PreviewItem struct {
	MovieName  string `json:"movie_name"`
	MovieJavID string `json:"movie_jav_id"`
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	CanMove    bool   `json:"can_move"`
	Error      string `json:"error"`
}

type PreviewResult struct {
	PlanID  string         `json:"plan_id"`
	Total   int            `json:"total"`
	Movable int            `json:"movable"`
	Failed  int            `json:"failed"`
	Items   []*PreviewItem `json:"items"`
}

type CommitItem struct {
	MovieName  string `json:"movie_name"`
	MovieJavID string `json:"movie_jav_id"`
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}

type CommitResult struct {
	PlanID       string        `json:"plan_id"`
	Total        int           `json:"total"`
	Success      int           `json:"success"`
	Failed       int           `json:"failed"`
	Items        []*CommitItem `json:"items"`
	SuccessItems []*CommitItem `json:"success_items"`
	FailedItems  []*CommitItem `json:"failed_items"`
}

type movePlan struct {
	ID        string
	CreatedAt int64
	Items     []*movePlanItem
}

type movePlanItem struct {
	MovieName  string
	MovieJavID string
	SourcePath string
	TargetPath string
	CanMove    bool
	Error      string
}

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps:     d,
		movieSvc: movie.NewService(d),
		plans:    make(map[string]*movePlan),
	}
}

func (s *Service) resolveTargetDir(rootDir string) string {
	rootDir = filepath.Clean(strings.TrimSpace(rootDir))
	if rootDir == "" {
		return ""
	}
	for _, pair := range s.deps.Config.Film.Pairs {
		if filepath.Clean(strings.TrimSpace(pair.RootDir)) != rootDir {
			continue
		}
		target := filepath.Clean(strings.TrimSpace(pair.MoveFilmDestination))
		if target != "" {
			return target
		}
		return ""
	}
	return ""
}

func (s *Service) nextPlanID() string {
	seq := atomic.AddUint64(&s.seq, 1)
	return fmt.Sprintf("film_move_%d_%d", time.Now().UnixNano(), seq)
}

func (s *Service) savePlan(items []*movePlanItem) string {
	now := time.Now().Unix()
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pruneExpiredPlansLocked(now)
	planID := s.nextPlanID()
	s.plans[planID] = &movePlan{
		ID:        planID,
		CreatedAt: now,
		Items:     items,
	}
	return planID
}

func (s *Service) takePlan(planID string) *movePlan {
	planID = strings.TrimSpace(planID)
	if planID == "" {
		return nil
	}

	now := time.Now().Unix()
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pruneExpiredPlansLocked(now)
	plan, ok := s.plans[planID]
	if !ok {
		return nil
	}
	delete(s.plans, planID)
	return plan
}

func (s *Service) pruneExpiredPlansLocked(now int64) {
	for id, plan := range s.plans {
		if plan == nil {
			delete(s.plans, id)
			continue
		}
		if now-plan.CreatedAt <= defaultPlanTTLSeconds {
			continue
		}
		delete(s.plans, id)
	}
}
