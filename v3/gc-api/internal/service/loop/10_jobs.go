package loop

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/service/fetchsehuatang"
	"rudy_gc/internal/service/sc"
	"rudy_gc/internal/taskctx"
)

const (
	TaskSpiderDailyBest              = "spider_daily_best"
	TaskSpiderDailyBestSync          = "spider_daily_best_sync"
	TaskSpiderSeeds                  = "spider_seeds"
	TaskSpiderSeedByName             = "spider_seed_by_name"
	TaskSpiderRefreshOldest          = "spider_refresh_oldest_detail"
	TaskSpiderDownloadCover          = "spider_download_cover"
	TaskSpiderTranslateTitle         = "spider_translate_title"
	TaskSpiderPostProcess            = "spider_post_process"
	TaskSpiderRebuildCastRank        = "spider_rebuild_cast_rank"
	TaskSpiderRebuildActorRank       = "spider_rebuild_actor_rank"
	TaskSpiderBackfillPerson         = "spider_backfill_person"
	TaskSpiderBackfillRankPeriod     = "spider_backfill_rank_period"
	TaskSpiderBackfillFetchSite      = "spider_backfill_fetch_site"
	TaskSpiderFetchJavbus            = "spider_fetch_javbus_resources"
	TaskSpiderFetchJavbusFilter      = "spider_fetch_javbus_filtered_resources"
	TaskSpiderFetchSukebei           = "spider_fetch_sukebei_resources"
	TaskSpiderFetchSukebeiFilter     = "spider_fetch_sukebei_filtered_resources"
	TaskSpiderFetchSiteBoth          = "spider_fetch_site_both_resources"
	TaskSpiderFetchSehuatangSH       = "spider_fetch_sehuatang_sh_threads"
	TaskSpiderFetchSehuatangSHList   = "spider_fetch_sehuatang_sh_thread_list"
	TaskSpiderFetchSehuatangSHDetail = "spider_fetch_sehuatang_sh_thread_detail"
	TaskScRebuildStats               = "sc_rebuild_stats"
	TaskScAdd                        = "sc_add"
)

const (
	taskGroupFetchPriority = "fetch_priority"
	taskGroupRefreshOldest = "refresh_oldest"
)

type StartTaskRequest struct {
	TaskType        string               `json:"task_type"`
	TriggerSource   string               `json:"trigger_source"`
	Name            string               `json:"name"`
	ActorName       string               `json:"actor_name"`
	AutoFetchSite   string               `json:"auto_fetch_site"`
	Keyword         string               `json:"keyword"`
	FetchMode       string               `json:"fetch_mode"`
	TargetForumIDs  []string             `json:"target_forum_ids"`
	ThreadTIDs      []string             `json:"thread_tids"`
	Sort            string               `json:"sort"`
	Status          string               `json:"status"`
	DetailStatus    string               `json:"detail_status"`
	DetailStatuses  []string             `json:"detail_statuses"`
	MovieType       string               `json:"movie_type"`
	MovieTypes      []string             `json:"movie_types"`
	Statuses        []string             `json:"statuses"`
	TriggerSort     string               `json:"trigger_sort"`
	TriggerOrder    string               `json:"trigger_order"`
	LastFetchFrom   string               `json:"last_fetch_from"`
	LastFetchTo     string               `json:"last_fetch_to"`
	PostTimeFrom    string               `json:"post_time_from"`
	PostTimeTo      string               `json:"post_time_to"`
	ReleaseDateFrom string               `json:"release_date_from"`
	ReleaseDateTo   string               `json:"release_date_to"`
	MediaBirthFrom  string               `json:"media_birth_from"`
	MediaBirthTo    string               `json:"media_birth_to"`
	StartPage       int64                `json:"start_page"`
	EndPage         int64                `json:"end_page"`
	Number          int64                `json:"number"`
	Limit           int64                `json:"limit"`
	MovieJavID      string               `json:"movie_jav_id"`
	MovieName       string               `json:"movie_name"`
	Dir             string               `json:"dir"`
	ComeMovieType   string               `json:"come_movie_type"`
	ComeMovieJavID  string               `json:"come_movie_jav_id"`
	ComeFC2MovieID  int64                `json:"come_fc2_movie_id"`
	MovieCast       string               `json:"movie_cast"`
	Kind            string               `json:"kind"`
	DurationMinutes int64                `json:"duration"`
	Fg              string               `json:"fg"`
	Vessel          string               `json:"vessel"`
	Remarks         string               `json:"remarks"`
	Movies          []sc.AddScInputMovie `json:"movies"`

	CastNames               string `json:"cn"`
	PersonIds               string `json:"pid"`
	GenreNames              string `json:"gn"`
	DirectorName            string `json:"dn"`
	PrefixName              string `json:"pn"`
	MakerName               string `json:"mn"`
	LabelName               string `json:"ln"`
	ReleasingDateStart      string `json:"rs"`
	ReleasingDateEnd        string `json:"re"`
	MediaBirthTimeStart     string `json:"mbs"`
	MediaBirthTimeEnd       string `json:"mbe"`
	CastAgeMin              string `json:"cay"`
	CastAgeMax              string `json:"cao"`
	StartRankingDateFrom    string `json:"srds"`
	StartRankingDateTo      string `json:"srde"`
	DaysInRankMin           string `json:"drkmin"`
	NeedDownload            string `json:"nd"`
	Word                    string `json:"wd"`
	MediaOwned              string `json:"mowned"`
	ViewWatchedMin          string `json:"vwmin"`
	ViewWatchedMax          string `json:"vwmax"`
	ScoreMin                string `json:"smin"`
	ScoreMax                string `json:"smax"`
	LastScTimeMin           string `json:"lsctmin"`
	LastScTimeMax           string `json:"lsctmax"`
	ScTimesMin              string `json:"scmin"`
	ScTimesMax              string `json:"scmax"`
	ComeTimesMin            string `json:"comin"`
	ComeTimesMax            string `json:"comax"`
	MediaDir1               string `json:"md1"`
	MediaDir2               string `json:"md2"`
	MediaDir3               string `json:"md3"`
	MediaDir4               string `json:"md4"`
	OrderBy                 string `json:"od"`
	Order                   string `json:"order"`
	LastFetchDurationDays   string `json:"last_fetch_duration_days"`
	LastSuccessDurationDays string `json:"last_success_duration_days"`
}

type DetailLoopSnapshot struct {
	Running     bool   `json:"running"`
	Paused      bool   `json:"paused"`
	Buffered    int    `json:"buffered"`
	StartedAt   int64  `json:"started_at"`
	LastEventAt int64  `json:"last_event_at"`
	LastLine    string `json:"last_line"`
}

func (l *FetchLoopService) Start(ctx context.Context) {
	l.rootMu.Lock()
	defer l.rootMu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}
	l.rootCtx = ctx
	if l.started {
		return
	}
	l.started = true
	l.startDetailLoopLocked(ctx, l.detailBaseWindow, l.detailMaxBatch)
}

func (l *FetchLoopService) Shutdown() {
	l.stopDetailLoop()
}

func (l *FetchLoopService) ListJobs() []JobSnapshot {
	return l.jobs.listRunning()
}

func (l *FetchLoopService) ListAllJobs() []JobSnapshot {
	return l.jobs.listAll()
}

func (l *FetchLoopService) GetJob(jobID string) (JobSnapshot, error) {
	snapshot, ok := l.jobs.snapshot(strings.TrimSpace(jobID))
	if !ok {
		return JobSnapshot{}, fmt.Errorf("job_id not found")
	}
	return snapshot, nil
}

func (l *FetchLoopService) GetJobEvents(jobID string, afterID int64) ([]JobEvent, error) {
	events, ok := l.jobs.history(strings.TrimSpace(jobID), afterID)
	if !ok {
		return nil, fmt.Errorf("job_id not found")
	}
	return events, nil
}

func (l *FetchLoopService) GetDetailLoopSnapshot() DetailLoopSnapshot {
	l.detailMu.Lock()
	snapshot := DetailLoopSnapshot{
		Running:   l.detailCancel != nil,
		Paused:    l.detailPaused,
		Buffered:  len(l.deps.DetailJobs),
		StartedAt: l.detailStartedAt,
	}
	l.detailMu.Unlock()

	if event, ok := l.detailLogs.latest(); ok {
		snapshot.LastEventAt = event.At
		snapshot.LastLine = event.Line
	}
	return snapshot
}

func (l *FetchLoopService) SubscribeJob(jobID string, afterID int64) ([]JobEvent, <-chan JobEvent, func(), error) {
	return l.jobs.subscribe(jobID, afterID)
}

func (l *FetchLoopService) SubscribeAllJobs(afterID int64) ([]JobEvent, <-chan JobEvent, func()) {
	return l.jobs.subscribeAll(afterID)
}

func (l *FetchLoopService) SubscribeDetailLoop() ([]DetailLoopEvent, <-chan DetailLoopEvent, func()) {
	return l.detailLogs.subscribe()
}

func (l *FetchLoopService) StartManagedDetailLoop() DetailLoopSnapshot {
	l.rootMu.Lock()
	rootCtx := l.rootCtx
	l.rootMu.Unlock()
	l.StartDetailLoop(rootCtx, l.detailBaseWindow, l.detailMaxBatch)
	return l.GetDetailLoopSnapshot()
}

func (l *FetchLoopService) StopManagedDetailLoop() DetailLoopSnapshot {
	l.StopDetailLoop()
	return l.GetDetailLoopSnapshot()
}

func (l *FetchLoopService) StopJob(jobID string) error {
	return l.jobs.stop(strings.TrimSpace(jobID))
}

func (l *FetchLoopService) PauseJob(jobID string) error {
	jobID = strings.TrimSpace(jobID)
	event, err := l.jobs.pause(jobID)
	if err != nil {
		return err
	}
	if event != nil {
		l.jobs.publish(jobID, *event)
	}
	return nil
}

func (l *FetchLoopService) ResumeJob(jobID string) error {
	jobID = strings.TrimSpace(jobID)
	if snapshot, ok := l.jobs.snapshot(jobID); ok {
		if snapshot.TaskType == TaskSpiderRefreshOldest && l.isExclusiveGroupRunning(taskGroupFetchPriority) {
			return fmt.Errorf("高优先抓取任务运行中，刷新最久详情暂不可继续")
		}
	}
	event, err := l.jobs.resume(jobID)
	if err != nil {
		return err
	}
	if event != nil {
		l.jobs.publish(jobID, *event)
	}
	return nil
}

func (l *FetchLoopService) StartTask(req StartTaskRequest) (string, error) {
	switch strings.TrimSpace(req.TaskType) {
	case TaskSpiderDailyBest:
		autoFetchSite, err := parseAutoFetchSiteEnabled(req.AutoFetchSite, true)
		if err != nil {
			return "", err
		}
		return l.StartDailyBest(autoFetchSite)
	case TaskSpiderDailyBestSync:
		autoFetchSite, err := parseAutoFetchSiteEnabled(req.AutoFetchSite, true)
		if err != nil {
			return "", err
		}
		return l.StartDailyBestSync(autoFetchSite)
	case TaskSpiderSeeds:
		autoFetchSite, err := parseAutoFetchSiteEnabled(req.AutoFetchSite, true)
		if err != nil {
			return "", err
		}
		return l.StartSeeds(autoFetchSite)
	case TaskSpiderSeedByName:
		name := strings.TrimSpace(req.Name)
		if name == "" {
			return "", fmt.Errorf("name is required")
		}
		autoFetchSite, err := parseAutoFetchSiteEnabled(req.AutoFetchSite, true)
		if err != nil {
			return "", err
		}
		return l.StartSeedByName(name, autoFetchSite)
	case TaskSpiderRefreshOldest:
		if req.Number <= 0 {
			return "", fmt.Errorf("number is required")
		}
		return l.StartRefreshOldestDetail(req.Number)
	case TaskSpiderDownloadCover:
		return l.StartDownloadCover()
	case TaskSpiderTranslateTitle:
		statuses, err := parseTranslateStatuses(req)
		if err != nil {
			return "", err
		}
		return l.StartTranslateTitle(statuses)
	case TaskSpiderPostProcess:
		statuses, err := parseTranslateStatuses(req)
		if err != nil {
			return "", err
		}
		return l.StartPostProcess(statuses)
	case TaskSpiderRebuildCastRank:
		return l.StartRebuildCastRank()
	case TaskSpiderRebuildActorRank:
		actorName := strings.TrimSpace(req.ActorName)
		if actorName == "" {
			return "", fmt.Errorf("actor_name is required")
		}
		return l.StartRebuildActorRank(actorName)
	case TaskSpiderBackfillPerson:
		return l.StartBackfillPerson()
	case TaskSpiderBackfillRankPeriod:
		return l.StartBackfillRankPeriod()
	case TaskSpiderBackfillFetchSite:
		return l.StartBackfillFetchSite()
	case TaskSpiderFetchJavbus:
		return l.StartFetchJavbusResources(req)
	case TaskSpiderFetchJavbusFilter:
		return l.StartFetchJavbusFilteredResources(req)
	case TaskSpiderFetchSukebei:
		return l.StartFetchSukebeiResources(req)
	case TaskSpiderFetchSukebeiFilter:
		return l.StartFetchSukebeiFilteredResources(req)
	case TaskSpiderFetchSiteBoth:
		return l.StartFetchSiteBothResources(req)
	case TaskSpiderFetchSehuatangSH:
		return l.StartFetchSehuatangSHThreads(req)
	case TaskSpiderFetchSehuatangSHList:
		return l.StartFetchSehuatangSHThreadList(req)
	case TaskSpiderFetchSehuatangSHDetail:
		targetForumIDs, err := normalizeSHDetailTaskTargetForumIDs(req.TriggerSource, req.TargetForumIDs)
		if err != nil {
			return "", err
		}
		req.TargetForumIDs = targetForumIDs
		return l.StartFetchSehuatangSHThreadDetails(req)
	case TaskScRebuildStats:
		return l.StartScRebuildStats()
	case TaskScAdd:
		dir := strings.TrimSpace(req.Dir)
		if dir == "" {
			return "", fmt.Errorf("dir is required")
		}
		return l.StartScAdd(sc.AddScInput{
			Dir:             dir,
			ComeMovieType:   strings.TrimSpace(req.ComeMovieType),
			ComeMovieJavId:  strings.TrimSpace(req.ComeMovieJavID),
			ComeFC2MovieId:  req.ComeFC2MovieID,
			MovieCast:       strings.TrimSpace(req.MovieCast),
			Kind:            strings.TrimSpace(req.Kind),
			DurationMinutes: req.DurationMinutes,
			Fg:              strings.TrimSpace(req.Fg),
			Vessel:          strings.TrimSpace(req.Vessel),
			Remarks:         strings.TrimSpace(req.Remarks),
			Movies:          req.Movies,
		})
	default:
		return "", fmt.Errorf("unsupported task_type: %s", req.TaskType)
	}
}

func normalizeSHDetailTaskTargetForumIDs(triggerSource string, values []string) ([]string, error) {
	normalized := uniqueNonEmpty(values)
	if strings.TrimSpace(triggerSource) == "shpool" {
		return normalized, nil
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("请选择目标版块")
	}
	return normalized, nil
}

func (l *FetchLoopService) StartDailyBest(autoFetchSite bool) (string, error) {
	return l.startManagedTask(TaskSpiderDailyBest, taskRuntimePolicy{
		ExclusiveGroup:        taskGroupFetchPriority,
		PauseDetailLoop:       true,
		PreemptsRefreshOldest: true,
	}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:           "pipeline_pre",
			Message:         "开始抓取每日榜",
			CurrentPhaseKey: "bestinv",
		})
		return l.crawlLogic.CrawlDailyBestProcession(ctx, false, autoFetchSite)
	})
}

func (l *FetchLoopService) StartDailyBestSync(autoFetchSite bool) (string, error) {
	return l.startManagedTask(TaskSpiderDailyBestSync, taskRuntimePolicy{
		ExclusiveGroup:        taskGroupFetchPriority,
		PauseDetailLoop:       true,
		PreemptsRefreshOldest: true,
	}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:           "pipeline_pre",
			Message:         "开始同步每日榜",
			CurrentPhaseKey: "bestinv",
		})
		return l.crawlLogic.CrawlDailyBestProcession(ctx, true, autoFetchSite)
	})
}

func (l *FetchLoopService) StartSeeds(autoFetchSite bool) (string, error) {
	return l.startManagedTask(TaskSpiderSeeds, taskRuntimePolicy{
		ExclusiveGroup:        taskGroupFetchPriority,
		PauseDetailLoop:       true,
		PreemptsRefreshOldest: true,
	}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: "开始抓取活跃种子"})
		return l.crawlLogic.CrawlBySeedsActiveProcession(ctx, autoFetchSite)
	})
}

func (l *FetchLoopService) StartSeedByName(name string, autoFetchSite bool) (string, error) {
	return l.startManagedTask(TaskSpiderSeedByName, taskRuntimePolicy{
		ExclusiveGroup:        taskGroupFetchPriority,
		PauseDetailLoop:       true,
		PreemptsRefreshOldest: true,
	}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: fmt.Sprintf("开始按名称抓取：%s", name)})
		return l.crawlLogic.CrawlBySeedName(ctx, name, autoFetchSite)
	})
}

func (l *FetchLoopService) StartRefreshOldestDetail(number int64) (string, error) {
	return l.startManagedTask(TaskSpiderRefreshOldest, taskRuntimePolicy{
		ExclusiveGroup:         taskGroupRefreshOldest,
		PauseDetailLoop:        true,
		RegistersRefreshOldest: true,
	}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: fmt.Sprintf("开始刷新最久未更新详情，数量=%d", number)})
		_, err := l.crawlLogic.RefreshOldestDetail(ctx, number)
		return err
	})
}

func (l *FetchLoopService) StartDownloadCover() (string, error) {
	return l.startManagedTask(TaskSpiderDownloadCover, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:           "pipeline_pre",
			Message:         "开始单独执行封面抓取",
			CurrentPhaseKey: "cover",
		})
		return l.crawlLogic.RunPostProcess(ctx, true, false, nil)
	})
}

func (l *FetchLoopService) StartTranslateTitle(statuses []int64) (string, error) {
	return l.startManagedTask(TaskSpiderTranslateTitle, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:           "pipeline_pre",
			Message:         fmt.Sprintf("开始单独执行标题翻译，状态=%s", joinTranslateStatuses(statuses)),
			CurrentPhaseKey: "translate",
		})
		return l.crawlLogic.RunPostProcess(ctx, false, true, statuses)
	})
}

func (l *FetchLoopService) StartPostProcess(statuses []int64) (string, error) {
	return l.startManagedTask(TaskSpiderPostProcess, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:           "pipeline_pre",
			Message:         fmt.Sprintf("开始执行封面抓取与标题翻译，翻译状态=%s", joinTranslateStatuses(statuses)),
			CurrentPhaseKey: "cover",
		})
		return l.crawlLogic.RunPostProcess(ctx, true, true, statuses)
	})
}

func parseTranslateStatuses(req StartTaskRequest) ([]int64, error) {
	rawValues := make([]string, 0, len(req.Statuses)+1)
	for _, value := range req.Statuses {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		rawValues = append(rawValues, trimmed)
	}
	if len(rawValues) == 0 {
		if single := strings.TrimSpace(req.Status); single != "" {
			rawValues = append(rawValues, single)
		}
	}
	if len(rawValues) == 0 {
		return []int64{consts.ItemChineseNone}, nil
	}

	allowed := map[int64]struct{}{
		consts.ItemChineseNone:      {},
		consts.ItemChineseOK:        {},
		consts.ItemChineseError:     {},
		consts.ItemChineseSensitive: {},
	}
	out := make([]int64, 0, len(rawValues))
	seen := make(map[int64]struct{}, len(rawValues))
	for _, raw := range rawValues {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("statuses 参数无效: %s", raw)
		}
		if _, ok := allowed[value]; !ok {
			return nil, fmt.Errorf("statuses 参数不支持: %d", value)
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return []int64{consts.ItemChineseNone}, nil
	}
	return out, nil
}

func joinTranslateStatuses(statuses []int64) string {
	if len(statuses) == 0 {
		return "1"
	}
	parts := make([]string, 0, len(statuses))
	for _, status := range statuses {
		parts = append(parts, strconv.FormatInt(status, 10))
	}
	return strings.Join(parts, ",")
}

func (l *FetchLoopService) StartRebuildCastRank() (string, error) {
	return l.startManagedTask(TaskSpiderRebuildCastRank, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: "开始回填演员 rank"})
		return l.crawlLogic.RebuildAllCastRankStats(ctx)
	})
}

func (l *FetchLoopService) StartRebuildActorRank(actorName string) (string, error) {
	return l.startManagedTask(TaskSpiderRebuildActorRank, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: fmt.Sprintf("开始回填演员 rank：%s", actorName)})
		return l.crawlLogic.RebuildCastRankStatsByName(ctx, actorName)
	})
}

func (l *FetchLoopService) StartBackfillPerson() (string, error) {
	return l.startManagedTask(TaskSpiderBackfillPerson, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: "开始回填 person"})
		return l.crawlLogic.BackfillPersonData(ctx)
	})
}

func (l *FetchLoopService) StartBackfillRankPeriod() (string, error) {
	return l.startManagedTask(TaskSpiderBackfillRankPeriod, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: "开始回填周期排行"})
		return l.movieSvc.RebuildAllRankPeriods(ctx)
	})
}

func (l *FetchLoopService) StartBackfillFetchSite() (string, error) {
	return l.startManagedTask(TaskSpiderBackfillFetchSite, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: "开始回填外站抓取任务"})
		result, err := l.fetchQueue.BackfillFetchTasks(ctx, 200)
		if err != nil {
			return err
		}
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:        "fetch_site_backfill_done",
			Message:      fmt.Sprintf("外站抓取任务回填完成：扫描=%d，新增=%d", result.Scanned, result.Created),
			HandledCount: int(result.Scanned),
			SuccessCount: int(result.Created),
		})
		return nil
	})
}

func (l *FetchLoopService) StartFetchJavbusResources(req StartTaskRequest) (string, error) {
	return l.startManagedTask(TaskSpiderFetchJavbus, taskRuntimePolicy{}, func(ctx context.Context) error {
		movieJavID := strings.TrimSpace(req.MovieJavID)
		movieName := strings.TrimSpace(req.MovieName)
		msg := "开始抓取 JavBus 资源"
		if movieJavID != "" {
			msg = fmt.Sprintf("开始抓取 JavBus 资源：%s", movieName)
			taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: msg})
			_, err := l.fetchQueue.RunSingleJavbusFetchTask(ctx, movieJavID, movieName)
			return err
		}
		movieReq, err := buildFetchSiteMovieRequest(req)
		if err != nil {
			return err
		}
		durationFilter, err := buildFetchSiteDurationFilter(req)
		if err != nil {
			return err
		}
		targetCount := normalizeFetchSiteNumber(req.Number)
		msg = fmt.Sprintf("开始抓取 JavBus 资源，目标数量=%d，排序=%s", targetCount, movieReq.OrderBy)
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: msg})
		_, err = l.runJavbusFetchTasksWithFilterTarget(
			ctx,
			movieReq,
			targetCount,
			durationFilter.LastFetchDurationDays,
			durationFilter.LastSuccessDurationDays,
		)
		return err
	})
}

func (l *FetchLoopService) StartFetchSukebeiResources(req StartTaskRequest) (string, error) {
	return l.startManagedTask(TaskSpiderFetchSukebei, taskRuntimePolicy{}, func(ctx context.Context) error {
		movieJavID := strings.TrimSpace(req.MovieJavID)
		movieName := strings.TrimSpace(req.MovieName)
		msg := "开始抓取 Sukebei 资源"
		if movieJavID != "" {
			msg = fmt.Sprintf("开始抓取 Sukebei 资源：%s", movieName)
			taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: msg})
			_, err := l.fetchQueue.RunSingleSukebeiFetchTask(ctx, movieJavID, movieName)
			return err
		}
		movieReq, err := buildFetchSiteMovieRequest(req)
		if err != nil {
			return err
		}
		durationFilter, err := buildFetchSiteDurationFilter(req)
		if err != nil {
			return err
		}
		targetCount := normalizeFetchSiteNumber(req.Number)
		msg = fmt.Sprintf("开始抓取 Sukebei 资源，目标数量=%d，排序=%s", targetCount, movieReq.OrderBy)
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: msg})
		_, err = l.runSukebeiFetchTasksWithFilterTarget(
			ctx,
			movieReq,
			targetCount,
			durationFilter.LastFetchDurationDays,
			durationFilter.LastSuccessDurationDays,
		)
		return err
	})
}

func (l *FetchLoopService) StartFetchSiteBothResources(req StartTaskRequest) (string, error) {
	return l.startManagedTask(TaskSpiderFetchSiteBoth, taskRuntimePolicy{}, func(ctx context.Context) error {
		movieJavID := strings.TrimSpace(req.MovieJavID)
		movieName := strings.TrimSpace(req.MovieName)
		msg := "开始同时抓取 JavBus + Sukebei 资源"
		if movieJavID != "" {
			msg = fmt.Sprintf("开始同时抓取 JavBus + Sukebei 资源：%s", movieName)
			taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: msg})
			if _, err := l.fetchQueue.RunSingleJavbusFetchTask(ctx, movieJavID, movieName); err != nil {
				return err
			}
			_, err := l.fetchQueue.RunSingleSukebeiFetchTask(ctx, movieJavID, movieName)
			return err
		}

		movieReq, err := buildFetchSiteMovieRequest(req)
		if err != nil {
			return err
		}
		durationFilter, err := buildFetchSiteDurationFilter(req)
		if err != nil {
			return err
		}
		targetCount := normalizeFetchSiteNumber(req.Number)
		msg = fmt.Sprintf("开始同时抓取 JavBus + Sukebei 资源，目标数量=%d，排序=%s", targetCount, movieReq.OrderBy)
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: msg})
		if _, err = l.runJavbusFetchTasksWithFilterTarget(
			ctx,
			movieReq,
			targetCount,
			durationFilter.LastFetchDurationDays,
			durationFilter.LastSuccessDurationDays,
		); err != nil {
			return err
		}
		_, err = l.runSukebeiFetchTasksWithFilterTarget(
			ctx,
			movieReq,
			targetCount,
			durationFilter.LastFetchDurationDays,
			durationFilter.LastSuccessDurationDays,
		)
		return err
	})
}

func (l *FetchLoopService) StartFetchJavbusFilteredResources(req StartTaskRequest) (string, error) {
	return l.startManagedTask(TaskSpiderFetchJavbusFilter, taskRuntimePolicy{}, func(ctx context.Context) error {
		query, err := buildJavbusPageQueryFromTask(req)
		if err != nil {
			return err
		}
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:   "pipeline_pre",
			Message: fmt.Sprintf("开始按列表筛选抓取 JavBus，排序=%s %s", query.Sort, strings.ToUpper(query.Order)),
		})
		tasks, err := l.fetchSiteSvc.ListJavbusFetchTasksByPageQuery(ctx, query)
		if err != nil {
			return err
		}
		return l.runJavbusFilteredQueue(ctx, tasks)
	})
}

func (l *FetchLoopService) StartFetchSukebeiFilteredResources(req StartTaskRequest) (string, error) {
	return l.startManagedTask(TaskSpiderFetchSukebeiFilter, taskRuntimePolicy{}, func(ctx context.Context) error {
		query, err := buildSukebeiPageQueryFromTask(req)
		if err != nil {
			return err
		}
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:   "pipeline_pre",
			Message: fmt.Sprintf("开始按列表筛选抓取 Sukebei，排序=%s %s", query.Sort, strings.ToUpper(query.Order)),
		})
		tasks, err := l.fetchSiteSvc.ListSukebeiFetchTasksByPageQuery(ctx, query)
		if err != nil {
			return err
		}
		return l.runSukebeiFilteredQueue(ctx, tasks)
	})
}

func (l *FetchLoopService) StartFetchSehuatangSHThreads(req StartTaskRequest) (string, error) {
	return l.startManagedTask(TaskSpiderFetchSehuatangSH, taskRuntimePolicy{}, func(ctx context.Context) error {
		startPage := req.StartPage
		endPage := req.EndPage
		if startPage <= 0 {
			startPage = fetchsehuatang.DefaultStartPage
		}
		if endPage <= 0 {
			endPage = fetchsehuatang.DefaultEndPage
		}

		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:             "pipeline_pre",
			Message:           fmt.Sprintf("开始色花堂一体抓取，先执行列表建池，页码=%d-%d，版块=%s", startPage, endPage, strings.Join(req.TargetForumIDs, ",")),
			CurrentPhaseKey:   "list",
			PhaseKey:          "list",
			PhaseHandledCount: 0,
		})

		listResult, err := l.fetchSehuatangSvc.FetchSHThreadList(ctx, fetchsehuatang.SHFetchRequest{
			FetchMode:      fetchsehuatang.SHFetchModeForum,
			TargetForumIDs: req.TargetForumIDs,
			StartPage:      startPage,
			EndPage:        endPage,
		})
		if err != nil {
			return err
		}
		listStats := fetchsehuatang.BuildSHProgressStats(listResult)
		listStatsText := fetchsehuatang.BuildSHProgressStatsText(listResult)
		detailTargets := append([]string(nil), listResult.DetailThreadTIDs...)
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:             "fetch_sehuatang_sh_detail_prepare",
			Message:           fmt.Sprintf("列表建池完成，准备补全详情，待处理=%d", len(detailTargets)),
			HandledCount:      int(listResult.MatchedCount),
			SuccessCount:      int(listResult.FetchedCount),
			FailedCount:       int(listResult.FailedCount),
			CurrentPhaseKey:   "detail",
			PhaseKey:          "detail",
			PhaseHandledCount: 0,
			PhaseTotalCount:   len(detailTargets),
			Stats:             listStats,
			StatsText:         listStatsText,
		})

		var detailResult *fetchsehuatang.SHFetchResult
		if len(detailTargets) > 0 {
			selectionCriteria, marshalErr := json.Marshal(map[string]any{
				"trigger_source":   "shfetchall",
				"target_forum_ids": req.TargetForumIDs,
				"detail_statuses":  []string{fetchsehuatang.SHDetailStatusPending, fetchsehuatang.SHDetailStatusFailed},
				"list_start_page":  startPage,
				"list_end_page":    endPage,
			})
			if marshalErr != nil {
				return marshalErr
			}
			detailResult, err = l.fetchSehuatangSvc.FetchSHThreadDetails(ctx, fetchsehuatang.SHDetailFetchRequest{
				TriggerSource:     "shfetchall",
				SelectionCriteria: string(selectionCriteria),
				TargetForumIDs:    req.TargetForumIDs,
				ThreadTIDs:        detailTargets,
				BaseStats:         listStats,
				BaseStatsText:     listStatsText,
			})
			if err != nil {
				return err
			}
		}

		stats := make(map[string]int64, len(listStats)+20)
		for key, value := range listStats {
			stats[key] = value
		}
		stats["detail_target_count"] = int64(len(detailTargets))
		statsText := make(map[string]string, len(listStatsText)+3)
		for key, value := range listStatsText {
			statsText[key] = value
		}
		detailMatched := int64(0)
		detailFetched := int64(0)
		detailFailed := int64(0)
		if detailResult != nil {
			detailMatched = detailResult.MatchedCount
			detailFetched = detailResult.FetchedCount
			detailFailed = detailResult.FailedCount
			for key, value := range fetchsehuatang.BuildSHProgressStats(detailResult) {
				stats["detail_"+key] = value
			}
			for key, value := range fetchsehuatang.BuildSHProgressStatsText(detailResult) {
				statsText["detail_"+key] = value
			}
		}

		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:             "fetch_sehuatang_sh_done",
			Message:           fmt.Sprintf("色花堂一体抓取完成：列表页面成功=%d，页面失败=%d，新增=%d，更新已有=%d，详情目标=%d，详情成功=%d，详情失败=%d", listResult.SuccessPageCount, listResult.FailedPageCount, listResult.InsertedCount, listResult.ExistingCount, detailMatched, detailFetched, detailFailed),
			HandledCount:      int(listResult.MatchedCount + detailMatched),
			SuccessCount:      int(listResult.FetchedCount + detailFetched),
			FailedCount:       int(listResult.FailedCount + detailFailed),
			CurrentPhaseKey:   "detail",
			PhaseKey:          "detail",
			PhaseHandledCount: int(detailMatched),
			PhaseTotalCount:   len(detailTargets),
			PhaseSuccessCount: int(detailFetched),
			PhaseFailedCount:  int(detailFailed),
			Stats:             stats,
			StatsText:         statsText,
		})
		return nil
	})
}

func (l *FetchLoopService) StartFetchSehuatangSHThreadList(req StartTaskRequest) (string, error) {
	return l.startManagedTask(TaskSpiderFetchSehuatangSHList, taskRuntimePolicy{}, func(ctx context.Context) error {
		startPage := req.StartPage
		endPage := req.EndPage
		if startPage <= 0 {
			startPage = fetchsehuatang.DefaultStartPage
		}
		if endPage <= 0 {
			endPage = fetchsehuatang.DefaultEndPage
		}

		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:   "pipeline_pre",
			Message: fmt.Sprintf("开始色花堂帖子列表建池，页码=%d-%d，版块=%s", startPage, endPage, strings.Join(req.TargetForumIDs, ",")),
		})

		result, err := l.fetchSehuatangSvc.FetchSHThreadList(ctx, fetchsehuatang.SHFetchRequest{
			FetchMode:      fetchsehuatang.SHFetchModeForum,
			TargetForumIDs: req.TargetForumIDs,
			StartPage:      startPage,
			EndPage:        endPage,
		})
		if err != nil {
			return err
		}

		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:        "fetch_sehuatang_sh_list_done",
			Message:      fmt.Sprintf("色花堂帖子列表抓取完成：run_id=%d，页面成功=%d，页面失败=%d，新增=%d，更新已有=%d，入库失败=%d，跳过置顶=%d", result.RunID, result.SuccessPageCount, result.FailedPageCount, result.InsertedCount, result.ExistingCount, result.PersistFailedCount, result.SkippedTopCount),
			HandledCount: int(result.MatchedCount),
			SuccessCount: int(result.FetchedCount),
			FailedCount:  int(result.FailedCount),
			Stats:        fetchsehuatang.BuildSHProgressStats(result),
			StatsText:    fetchsehuatang.BuildSHProgressStatsText(result),
		})
		return nil
	})
}

func (l *FetchLoopService) StartFetchSehuatangSHThreadDetails(req StartTaskRequest) (string, error) {
	return l.startManagedTask(TaskSpiderFetchSehuatangSHDetail, taskRuntimePolicy{}, func(ctx context.Context) error {
		selectionCriteria, err := json.Marshal(map[string]any{
			"source_forum_ids": req.TargetForumIDs,
			"detail_statuses":  req.DetailStatuses,
			"movie_types":      req.MovieTypes,
			"keyword":          strings.TrimSpace(req.Keyword),
			"post_time_from":   strings.TrimSpace(req.PostTimeFrom),
			"post_time_to":     strings.TrimSpace(req.PostTimeTo),
		})
		if err != nil {
			return err
		}
		request := fetchsehuatang.SHDetailFetchRequest{
			TriggerSource:     req.TriggerSource,
			SelectionCriteria: string(selectionCriteria),
			TargetForumIDs:    req.TargetForumIDs,
			Keyword:           req.Keyword,
			ThreadTIDs:        req.ThreadTIDs,
		}

		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:   "pipeline_pre",
			Message: fmt.Sprintf("开始抓取选中帖子详情，选中=%d", len(request.ThreadTIDs)),
		})

		result, err := l.fetchSehuatangSvc.FetchSHThreadDetails(ctx, request)
		if err != nil {
			return err
		}

		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:        "fetch_sehuatang_sh_detail_done",
			Message:      fmt.Sprintf("色花堂选中帖子详情抓取完成：选中=%d，成功=%d，失败=%d", result.MatchedCount, result.FetchedCount, result.FailedCount),
			HandledCount: int(result.MatchedCount),
			SuccessCount: int(result.FetchedCount),
			FailedCount:  int(result.FailedCount),
			Stats:        fetchsehuatang.BuildSHProgressStats(result),
			StatsText:    fetchsehuatang.BuildSHProgressStatsText(result),
		})
		return nil
	})
}

func (l *FetchLoopService) StartScRebuildStats() (string, error) {
	return l.startManagedTask(TaskScRebuildStats, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "sc_prepare", Message: "开始回填 SC 统计"})
		return l.scSvc.RebuildAllScStats(ctx)
	})
}

func (l *FetchLoopService) StartScAdd(in sc.AddScInput) (string, error) {
	return l.startManagedTask(TaskScAdd, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "sc_prepare", Message: fmt.Sprintf("开始新增 SC：%s", in.Dir)})
		return l.scSvc.AddSc(ctx, in)
	})
}

func (l *FetchLoopService) startManagedTask(taskType string, policy taskRuntimePolicy, run func(context.Context) error) (string, error) {
	if run == nil {
		return "", fmt.Errorf("nil task runner")
	}
	jobID := l.jobs.create(taskType)
	if err := l.beginTaskRuntime(jobID, taskType, policy); err != nil {
		l.jobs.finish(jobID, JobEvent{
			Kind:     JobEventKindProgress,
			JobID:    jobID,
			TaskType: taskType,
			Stage:    "failed",
			Message:  err.Error(),
			Done:     true,
			At:       time.Now().Unix(),
		})
		return "", err
	}

	ctx, cancel := context.WithCancel(l.currentRootContext())
	if err := l.jobs.setCancel(jobID, cancel); err != nil {
		cancel()
		l.finishTaskRuntime(jobID, policy)
		return "", err
	}

	go func() {
		defer cancel()
		defer l.jobs.clearCancel(jobID)
		defer l.finishTaskRuntime(jobID, policy)

		ctx = taskctx.WithJobID(ctx, jobID)
		ctx = taskctx.WithPauseWaiter(ctx, func(ctx context.Context) error {
			return l.jobs.waitIfPaused(ctx, jobID)
		})
		ctx = taskctx.WithProgressReporter(ctx, func(progress taskctx.Progress) {
			l.publishTaskProgress(jobID, taskType, progress)
		})
		ctx = taskctx.WithLogReporter(ctx, func(logEvent taskctx.Log) {
			l.publishTaskLog(jobID, taskType, logEvent)
		})

		l.publishTaskProgress(jobID, taskType, taskctx.Progress{
			Stage:   "job_started",
			Message: "任务已启动",
		})

		l.applyTaskStartPolicy(jobID, policy)

		if err := l.jobs.waitIfPaused(ctx, jobID); err != nil {
			l.jobs.finish(jobID, l.buildManagedTaskFinalEvent(jobID, taskType, "failed", err.Error()))
			return
		}

		detailPauseApplied := false
		detailPauseStateChanged := false
		if policy.PauseDetailLoop && l.IsDetailLoopRunning() {
			detailPauseApplied, detailPauseStateChanged = l.pauseDetailLoop(jobID)
		}
		if detailPauseStateChanged {
			l.publishTaskProgress(jobID, taskType, taskctx.Progress{
				Stage:   "detail_paused",
				Message: "详情抓取已暂停",
			})
		}

		err := run(ctx)

		if detailPauseApplied {
			_, detailResumed := l.resumeDetailLoop(jobID)
			if detailResumed {
				l.publishTaskProgress(jobID, taskType, taskctx.Progress{
					Stage:   "detail_resumed",
					Message: "详情抓取已恢复",
				})
			}
		}
		if policy.PreemptsRefreshOldest {
			l.resumeRefreshOldestIfAutoPaused()
		}

		if err != nil {
			l.jobs.finish(jobID, l.buildManagedTaskFinalEvent(jobID, taskType, "failed", err.Error()))
			return
		}

		l.jobs.finish(jobID, l.buildManagedTaskFinalEvent(jobID, taskType, "done", "任务完成"))
	}()

	return jobID, nil
}
