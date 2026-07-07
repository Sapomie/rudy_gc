(function (global) {
    const API_BASE = '/api/crawler/jobs';
    const TASK_FETCH_SITE_BOTH = 'spider_fetch_site_both_resources';
    const TASK_FETCH_JAVBUS_FILTERED = 'spider_fetch_javbus_filtered_resources';
    const TASK_FETCH_SUKEBEI_FILTERED = 'spider_fetch_sukebei_filtered_resources';
    const TASK_FETCH_SEHUATANG = 'spider_fetch_sehuatang_magnets';
    const TASK_FETCH_SEHUATANG_SH = 'spider_fetch_sehuatang_sh_threads';
    const TASK_FETCH_SEHUATANG_SH_LIST = 'spider_fetch_sehuatang_sh_thread_list';
    const TASK_FETCH_SEHUATANG_SH_DETAIL = 'spider_fetch_sehuatang_sh_thread_detail';
    const FETCH_SITE_TASK_TYPES = ['spider_fetch_javbus_resources', 'spider_fetch_sukebei_resources'];
    const FETCH_SITE_DEFAULT_NUMBER = 1000000;
    const FETCH_SITE_DEFAULT_DURATION_DAYS = 5;
    const TASK_LABELS = {
        spider_daily_best: 'DailyBest 抓取',
        spider_daily_best_sync: 'DailyBest 同步',
        spider_seeds: '活跃 Seeds',
        spider_seed_by_name: '按名称抓取',
        spider_refresh_oldest_detail: '刷新最久详情',
        spider_download_cover: '图片抓取',
        spider_translate_title: '标题翻译',
        spider_post_process: '图片 + 标题翻译',
        spider_backfill_person: 'person 回填',
        spider_backfill_rank_period: '周期排行回填',
        spider_backfill_fetch_site: '外站任务回填',
        spider_fetch_javbus_resources: 'JavBus 资源抓取',
        spider_fetch_javbus_filtered_resources: 'JavBus 列表筛选抓取',
        spider_fetch_sukebei_resources: 'Sukebei 资源抓取',
        spider_fetch_sukebei_filtered_resources: 'Sukebei 列表筛选抓取',
        spider_fetch_site_both_resources: 'JavBus + Sukebei 同时抓取',
        spider_fetch_sehuatang_magnets: '色花堂磁力抓取',
        spider_fetch_sehuatang_sh_threads: '色花堂帖子新流程',
        spider_fetch_sehuatang_sh_thread_list: '色花堂帖子列表抓取',
        spider_fetch_sehuatang_sh_thread_detail: '色花堂帖子详情补全',
        spider_rebuild_cast_rank: '演员 Rank 回填',
        spider_rebuild_actor_rank: '单演员 Rank',
        sc_rebuild_stats: 'SC 统计回填',
        sc_move: 'SC 影片移动',
        sc_add: 'SC 新增',
    };

    const EVENT_LIMIT = 300;
    const REQUIRED_FIELDS = {
        spider_seed_by_name: 'name',
        spider_refresh_oldest_detail: 'number',
        spider_fetch_javbus_resources: 'number',
        spider_fetch_sukebei_resources: 'number',
        spider_fetch_site_both_resources: 'number',
        spider_rebuild_actor_rank: 'actor_name',
    };
    const MOVIE_FILTER_FIELD_NAMES = [
        'cn', 'pid', 'gn', 'dn', 'pn', 'mn', 'ln',
        'rs', 're', 'mbs', 'mbe',
        'cay', 'cao',
        'srds', 'srde',
        'drkmin', 'nd', 'wd', 'mowned',
        'vwmin', 'vwmax', 'smin', 'smax',
        'lsctmin', 'lsctmax', 'scmin', 'scmax', 'comin', 'comax',
        'md1', 'md2', 'md3', 'md4',
        'od', 'order',
    ];
    const FETCH_SITE_DURATION_FIELD_NAMES = [
        'last_fetch_duration_days',
        'last_success_duration_days',
    ];
    const STAGE_OVERVIEW_CONFIGS = {
        dailybest_stages: {
            phaseKeys: ['bestinv', 'detail', 'cover', 'translate'],
            labels: {
                bestinv: '榜单抓取',
                detail: '详情抓取',
                fetch_javbus: 'JavBus 抓取',
                fetch_sukebei: 'Sukebei 抓取',
                cover: '封面下载',
                translate: '标题翻译',
            },
            stageByEvent: {
                pipeline_pre: 'bestinv',
                job_started: 'bestinv',
                bestinv_prepare: 'bestinv',
                bestinv_page_done: 'bestinv',
                detail_queue_ready: 'detail',
                detail_item_done: 'detail',
                detail_fetch_failed: 'detail',
                detail_parse_failed: 'detail',
                fetch_site_after_detail_prepare: 'detail',
                fetch_javbus_queue_ready: 'fetch_javbus',
                fetch_javbus_done: 'fetch_javbus',
                fetch_javbus_failed: 'fetch_javbus',
                fetch_javbus_all_done: 'fetch_javbus',
                fetch_sukebei_queue_ready: 'fetch_sukebei',
                fetch_sukebei_done: 'fetch_sukebei',
                fetch_sukebei_failed: 'fetch_sukebei',
                fetch_sukebei_all_done: 'fetch_sukebei',
                cover_queue_ready: 'cover',
                cover_download_done: 'cover',
                cover_download_failed: 'cover',
                translate_queue_ready: 'translate',
                translate_done: 'translate',
                translate_failed: 'translate',
            },
            pageStylePhases: {
                bestinv: true,
            },
        },
        seeds_stages: {
            phaseKeys: ['seed', 'detail', 'cover', 'translate'],
            labels: {
                seed: '种子抓取',
                detail: '详情抓取',
                fetch_javbus: 'JavBus 抓取',
                fetch_sukebei: 'Sukebei 抓取',
                cover: '封面下载',
                translate: '标题翻译',
            },
            stageByEvent: {
                pipeline_pre: 'seed',
                job_started: 'seed',
                seed_queue_ready: 'seed',
                seed_item_done: 'seed',
                seed_page_done: 'seed',
                detail_queue_ready: 'detail',
                detail_item_done: 'detail',
                detail_fetch_failed: 'detail',
                detail_parse_failed: 'detail',
                fetch_site_after_detail_prepare: 'detail',
                fetch_javbus_queue_ready: 'fetch_javbus',
                fetch_javbus_done: 'fetch_javbus',
                fetch_javbus_failed: 'fetch_javbus',
                fetch_javbus_all_done: 'fetch_javbus',
                fetch_sukebei_queue_ready: 'fetch_sukebei',
                fetch_sukebei_done: 'fetch_sukebei',
                fetch_sukebei_failed: 'fetch_sukebei',
                fetch_sukebei_all_done: 'fetch_sukebei',
                cover_queue_ready: 'cover',
                cover_download_done: 'cover',
                cover_download_failed: 'cover',
                translate_queue_ready: 'translate',
                translate_done: 'translate',
                translate_failed: 'translate',
            },
            pageStylePhases: {},
        },
        post_process_stages: {
            phaseKeys: ['cover', 'translate'],
            labels: {
                cover: '封面下载',
                translate: '标题翻译',
            },
            stageByEvent: {
                pipeline_pre: 'cover',
                job_started: 'cover',
                cover_queue_ready: 'cover',
                cover_download_done: 'cover',
                cover_download_failed: 'cover',
                translate_queue_ready: 'translate',
                translate_done: 'translate',
                translate_failed: 'translate',
            },
            pageStylePhases: {},
        },
    };

    function createCrawlerJobConsole(config) {
        const page = document.getElementById('crawlerJobsPage');
        if (!page) {
            return null;
        }

        const runtime = {
            config: config || {},
            state: {
                selectedJobId: '',
                selectedJob: null,
                jobs: [],
                fetchSiteBothJobIds: [],
                fetchSitePreviewPayload: null,
                fetchSitePreviewPage: 1,
                fetchSitePreviewTotalPages: 0,
                fetchSitePreviewPageSize: 100,
                shDetailPreviewRows: [],
                detailLoop: null,
                eventEntries: [],
                eventCount: 0,
                lastEventId: 0,
                eventSource: null,
                eventStreamKey: '',
                reconnectTimer: null,
                loadTimer: null,
                elapsedTimer: null,
                fetchSiteFilterExpanded: false,
                destroyed: false,
            },
        };

        runtime.storageKey = runtime.config.storageKey || 'crawler_jobs_selected_job';
        runtime.defaultTaskType = String(runtime.config.defaultTaskType || '').trim();
        runtime.emptyStateText = String(runtime.config.emptyStateText || '等待触发').trim() || '等待触发';
        runtime.overviewExtraMode = String(runtime.config.overviewExtraMode || 'detail_loop').trim() || 'detail_loop';
        runtime.stageOverviewConfig = STAGE_OVERVIEW_CONFIGS[runtime.overviewExtraMode] || null;
        runtime.allowedTaskTypes = Array.isArray(runtime.config.allowedTaskTypes)
            ? runtime.config.allowedTaskTypes.map(function (item) {
                return String(item || '').trim();
            }).filter(function (item) {
                return item !== '';
            })
            : [];

        runtime.msgArea = document.getElementById('msgArea');
        runtime.form = document.getElementById('crawlerTaskForm');
        runtime.taskTypeInput = document.getElementById('task_type');
        runtime.submitBtn = document.getElementById('submitBtn');
        runtime.pauseBtn = document.getElementById('pauseBtn');
        runtime.resumeBtn = document.getElementById('resumeBtn');
        runtime.stopBtn = document.getElementById('stopBtn');
        runtime.debugRaw = document.getElementById('debugRaw');
        runtime.resultBox = document.getElementById('resultBox');
        runtime.progressBar = document.getElementById('progressBar');
        runtime.statusMsg = document.getElementById('statusMsg');
        runtime.eventCountText = document.getElementById('eventCountText');
        runtime.currentStageText = document.getElementById('currentStageText');
        runtime.detailLoopText = document.getElementById('detailLoopText');
        runtime.resultText = document.getElementById('resultText');
        runtime.elapsedText = document.getElementById('elapsedText');
        runtime.fetchSiteJavbusProgressText = document.getElementById('fetchSiteJavbusProgressText');
        runtime.fetchSiteJavbusResultText = document.getElementById('fetchSiteJavbusResultText');
        runtime.fetchSiteSukebeiProgressText = document.getElementById('fetchSiteSukebeiProgressText');
        runtime.fetchSiteSukebeiResultText = document.getElementById('fetchSiteSukebeiResultText');
        runtime.sehuatangProgressText = document.getElementById('sehuatangProgressText');
        runtime.sehuatangPageProgressText = document.getElementById('sehuatangPageProgressText');
        runtime.sehuatangCurrentPageText = document.getElementById('sehuatangCurrentPageText');
        runtime.sehuatangSeenText = document.getElementById('sehuatangSeenText');
        runtime.sehuatangMatchedText = document.getElementById('sehuatangMatchedText');
        runtime.sehuatangSkippedForumText = document.getElementById('sehuatangSkippedForumText');
        runtime.sehuatangSkippedExistingText = document.getElementById('sehuatangSkippedExistingText');
        runtime.sehuatangFetchedText = document.getElementById('sehuatangFetchedText');
        runtime.sehuatangFailedText = document.getElementById('sehuatangFailedText');
        runtime.filteredQueueTotalText = document.getElementById('filteredQueueTotalText') || document.getElementById('sukebeiFilteredTotalText');
        runtime.filteredQueuePendingText = document.getElementById('filteredQueuePendingText') || document.getElementById('sukebeiFilteredPendingText');
        runtime.filteredQueueSuccessText = document.getElementById('filteredQueueSuccessText') || document.getElementById('sukebeiFilteredSuccessText');
        runtime.filteredQueueFailedText = document.getElementById('filteredQueueFailedText') || document.getElementById('sukebeiFilteredFailedText');
        runtime.filteredQueuePendingList = document.getElementById('filteredQueuePendingList') || document.getElementById('sukebeiFilteredPendingList');
        runtime.filteredQueueDoneList = document.getElementById('filteredQueueDoneList') || document.getElementById('sukebeiFilteredDoneList');
        runtime.dailyBestBestinvProgressText = document.getElementById('dailyBestBestinvProgressText');
        runtime.dailyBestBestinvResultText = document.getElementById('dailyBestBestinvResultText');
        runtime.dailyBestDetailProgressText = document.getElementById('dailyBestDetailProgressText');
        runtime.dailyBestDetailResultText = document.getElementById('dailyBestDetailResultText');
        runtime.dailyBestCoverProgressText = document.getElementById('dailyBestCoverProgressText');
        runtime.dailyBestCoverResultText = document.getElementById('dailyBestCoverResultText');
        runtime.dailyBestTranslateProgressText = document.getElementById('dailyBestTranslateProgressText');
        runtime.dailyBestTranslateResultText = document.getElementById('dailyBestTranslateResultText');
        runtime.dailyBestFetchJavbusProgressText = document.getElementById('dailyBestFetchJavbusProgressText');
        runtime.dailyBestFetchJavbusResultText = document.getElementById('dailyBestFetchJavbusResultText');
        runtime.dailyBestFetchSukebeiProgressText = document.getElementById('dailyBestFetchSukebeiProgressText');
        runtime.dailyBestFetchSukebeiResultText = document.getElementById('dailyBestFetchSukebeiResultText');
        runtime.dailyBestStageCards = page.querySelectorAll('.overview-stage-card[data-phase]');
        runtime.nameFieldWrap = document.getElementById('nameFieldWrap');
        runtime.numberFieldWrap = document.getElementById('numberFieldWrap');
        runtime.actorNameFieldWrap = document.getElementById('actorNameFieldWrap');
        runtime.translateStatusesFieldWrap = document.getElementById('translateStatusesFieldWrap');
        runtime.autoFetchSiteFieldWrap = document.getElementById('autoFetchSiteFieldWrap');
        runtime.fetchSehuatangListURLFieldWrap = document.getElementById('fetchSehuatangListURLFieldWrap');
        runtime.fetchSehuatangKeywordFieldWrap = document.getElementById('fetchSehuatangKeywordFieldWrap');
        runtime.fetchSehuatangStopOnExistingFieldWrap = document.getElementById('fetchSehuatangStopOnExistingFieldWrap');
        runtime.fetchSehuatangStartPageFieldWrap = document.getElementById('fetchSehuatangStartPageFieldWrap');
        runtime.fetchSehuatangEndPageFieldWrap = document.getElementById('fetchSehuatangEndPageFieldWrap');
        runtime.fetchSehuatangPersistModeFieldWrap = document.getElementById('fetchSehuatangPersistModeFieldWrap');
        runtime.fetchSehuatangSHModeFieldWrap = document.getElementById('fetchSehuatangSHModeFieldWrap');
        runtime.fetchSehuatangSHForumsFieldWrap = document.getElementById('fetchSehuatangSHForumsFieldWrap');
        runtime.fetchSehuatangSHDetailStatusFieldWrap = document.getElementById('fetchSehuatangSHDetailStatusFieldWrap');
        runtime.fetchSehuatangSHDetailTitleFieldWrap = document.getElementById('fetchSehuatangSHDetailTitleFieldWrap');
        runtime.fetchSehuatangSHPostTimeFromFieldWrap = document.getElementById('fetchSehuatangSHPostTimeFromFieldWrap');
        runtime.fetchSehuatangSHPostTimeToFieldWrap = document.getElementById('fetchSehuatangSHPostTimeToFieldWrap');
        runtime.fetchSehuatangSHDetailLimitFieldWrap = document.getElementById('fetchSehuatangSHDetailLimitFieldWrap');
        runtime.fetchSiteFilterPanel = document.getElementById('fetchSiteFilterPanel');
        runtime.toggleFetchSiteFilterBtn = document.getElementById('toggleFetchSiteFilterBtn');
        runtime.fetchSiteFilterWrap = document.getElementById('fetchSiteFilterWrap');
        runtime.clearFetchSiteFilterBtn = document.getElementById('clearFetchSiteFilterBtn');
        runtime.previewFetchSiteBtn = document.getElementById('previewFetchSiteBtn');
        runtime.fetchSitePreviewPanel = document.getElementById('fetchSitePreviewPanel');
        runtime.fetchSitePreviewSummary = document.getElementById('fetchSitePreviewSummary');
        runtime.fetchSitePreviewMetaText = document.getElementById('fetchSitePreviewMetaText');
        runtime.fetchSitePreviewTableBody = document.getElementById('fetchSitePreviewTableBody');
        runtime.fetchSitePreviewPageText = document.getElementById('fetchSitePreviewPageText');
        runtime.fetchSitePreviewPrevBtn = document.getElementById('fetchSitePreviewPrevBtn');
        runtime.fetchSitePreviewNextBtn = document.getElementById('fetchSitePreviewNextBtn');
        runtime.hideFetchSitePreviewBtn = document.getElementById('hideFetchSitePreviewBtn');
        runtime.nameInput = document.getElementById('task_name');
        runtime.numberInput = document.getElementById('task_number');
        runtime.actorNameInput = document.getElementById('task_actor_name');
        runtime.autoFetchSiteInput = document.getElementById('task_auto_fetch_site');
        runtime.fetchSehuatangListURLInput = document.getElementById('task_sehuatang_list_url');
        runtime.fetchSehuatangKeywordInput = document.getElementById('task_sehuatang_keyword');
        runtime.fetchSehuatangStopOnExistingInput = document.getElementById('task_sehuatang_stop_on_existing');
        runtime.fetchSehuatangStartPageInput = document.getElementById('task_sehuatang_start_page');
        runtime.fetchSehuatangEndPageInput = document.getElementById('task_sehuatang_end_page');
        runtime.fetchSehuatangPersistModeInput = document.getElementById('task_sehuatang_persist_mode');
        runtime.fetchSehuatangSHModeInput = document.getElementById('task_sehuatang_sh_fetch_mode');
        runtime.fetchSehuatangSHDetailKeywordInput = document.getElementById('task_sehuatang_sh_detail_keyword');
        runtime.fetchSehuatangSHPostTimeFromInput = document.getElementById('task_sehuatang_sh_post_time_from');
        runtime.fetchSehuatangSHPostTimeToInput = document.getElementById('task_sehuatang_sh_post_time_to');
        runtime.fetchSehuatangSHDetailLimitInput = document.getElementById('task_sehuatang_sh_detail_limit');
        runtime.previewSHDetailBtn = document.getElementById('previewSHDetailBtn');
        runtime.shDetailPreviewPanel = document.getElementById('shDetailPreviewPanel');
        runtime.shDetailPreviewSummary = document.getElementById('shDetailPreviewSummary');
        runtime.shDetailPreviewMetaText = document.getElementById('shDetailPreviewMetaText');
        runtime.shDetailPreviewTableBody = document.getElementById('shDetailPreviewTableBody');
        runtime.selectAllSHDetailPreviewBtn = document.getElementById('selectAllSHDetailPreviewBtn');
        runtime.clearSHDetailPreviewBtn = document.getElementById('clearSHDetailPreviewBtn');
        runtime.hideSHDetailPreviewBtn = document.getElementById('hideSHDetailPreviewBtn');

        runtime.escapeHtml = function (input) {
            return String(input || '')
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/"/g, '&quot;')
                .replace(/'/g, '&#39;');
        };

        runtime.pickValue = function (value, fallback) {
            return value === undefined || value === null ? fallback : value;
        };

        runtime.pickNumber = function (value, fallback) {
            const picked = runtime.pickValue(value, fallback);
            return Number(picked || 0);
        };

        runtime.normalizeQueueItems = function (items) {
            if (!Array.isArray(items)) {
                return [];
            }
            return items.map(function (item) {
                return {
                    movie_jav_id: String(item && item.movie_jav_id || '').trim(),
                    movie_name: String(item && item.movie_name || '').trim(),
                    seq: Number(item && item.seq || 0),
                    state: String(item && item.state || '').trim(),
                    error: String(item && item.error || '').trim(),
                    completed_at: Number(item && item.completed_at || 0),
                };
            }).filter(function (item) {
                return item.movie_jav_id !== '' || item.movie_name !== '';
            });
        };

        runtime.normalizeJobId = function (value) {
            const raw = String(value || '').trim();
            if (!raw || raw === '""' || raw === "''") {
                return '';
            }
            if (!/^[A-Za-z0-9_.:-]+$/.test(raw)) {
                return '';
            }
            return raw;
        };

        runtime.isVisibleTaskType = function (taskType) {
            if (!runtime.allowedTaskTypes.length) {
                return true;
            }
            const current = String(taskType || '').trim();
            if (!current) {
                return false;
            }
            return runtime.allowedTaskTypes.indexOf(current) >= 0;
        };

        runtime.filterVisibleJobs = function (list) {
            const safeList = Array.isArray(list) ? list : [];
            return safeList.filter(function (item) {
                return item && runtime.isVisibleTaskType(item.task_type);
            });
        };

        runtime.timeText = function (ts) {
            const stamp = Number(ts || 0);
            if (!stamp) {
                return '-';
            }
            const d = new Date(stamp * 1000);
            if (Number.isNaN(d.getTime())) {
                return '-';
            }
            const y = d.getFullYear();
            const m = String(d.getMonth() + 1).padStart(2, '0');
            const day = String(d.getDate()).padStart(2, '0');
            const hh = String(d.getHours()).padStart(2, '0');
            const mm = String(d.getMinutes()).padStart(2, '0');
            const ss = String(d.getSeconds()).padStart(2, '0');
            return y + '-' + m + '-' + day + ' ' + hh + ':' + mm + ':' + ss;
        };

        runtime.formatElapsed = function (startedAt) {
            const start = Number(startedAt || 0);
            if (!start) {
                return '00:00';
            }
            const delta = Math.max(0, Math.floor(Date.now() / 1000) - start);
            const hour = Math.floor(delta / 3600);
            const minute = Math.floor((delta % 3600) / 60);
            const second = delta % 60;
            if (hour > 0) {
                return String(hour).padStart(2, '0') + ':' + String(minute).padStart(2, '0') + ':' + String(second).padStart(2, '0');
            }
            return String(minute).padStart(2, '0') + ':' + String(second).padStart(2, '0');
        };

        runtime.taskLabel = function (taskType) {
            return TASK_LABELS[String(taskType || '').trim()] || String(taskType || '-');
        };

        runtime.stageLabel = function (stage) {
            const current = String(stage || '').trim();
            switch (current) {
                case 'job_started':
                    return '任务已启动';
                case 'detail_paused':
                    return '详情抓取已暂停';
                case 'detail_resumed':
                    return '详情抓取已恢复';
                case 'paused':
                    return '任务已暂停';
                case 'resumed':
                    return '任务已继续';
                case 'done':
                    return '任务完成';
                case 'failed':
                    return '任务失败';
                default:
                    return current || runtime.emptyStateText;
            }
        };

        runtime.progressPercent = function (job) {
            const handled = Number((job || {}).handled_count || 0);
            const queued = Number((job || {}).queued_count || 0);
            const total = handled + queued;
            if (!total) {
                return 0;
            }
            return Math.max(0, Math.min(100, Math.round(handled * 100 / total)));
        };

        runtime.pageProgressText = function (job) {
            const handled = Number((job || {}).handled_count || 0);
            const queued = Number((job || {}).queued_count || 0);
            const total = handled + queued;
            if (total <= 0 || handled <= 0) {
                return '-';
            }
            return '第 ' + handled + '/' + total + ' 页';
        };

        runtime.jobStats = function (job) {
            if (!job || !job.stats || typeof job.stats !== 'object') {
                return {};
            }
            return job.stats;
        };

        runtime.jobStatsText = function (job) {
            if (!job || !job.stats_text || typeof job.stats_text !== 'object') {
                return {};
            }
            return job.stats_text;
        };

        runtime.jobStatNumber = function (job, key) {
            return Number(runtime.jobStats(job)[key] || 0);
        };

        runtime.hasJobStat = function (job, key) {
            return Object.prototype.hasOwnProperty.call(runtime.jobStats(job), key);
        };

        runtime.jobStatText = function (job, key) {
            return String(runtime.jobStatsText(job)[key] || '').trim();
        };

        runtime.stageOverviewPhaseKeyFromStage = function (stage) {
            if (!runtime.stageOverviewConfig || !runtime.stageOverviewConfig.stageByEvent) {
                return '';
            }
            return runtime.stageOverviewConfig.stageByEvent[String(stage || '').trim()] || '';
        };

        runtime.stageOverviewCurrentStageKey = function (job) {
            if (!job) {
                return '';
            }
            const phaseKey = String(job.current_phase_key || '').trim();
            if (phaseKey) {
                return phaseKey;
            }
            return runtime.stageOverviewPhaseKeyFromStage(job.stage);
        };

        runtime.stageOverviewPhaseStats = function (job) {
            if (!job || !job.phase_stats || typeof job.phase_stats !== 'object') {
                return {};
            }
            return job.phase_stats;
        };

        runtime.stageOverviewStageStat = function (job, stageKey) {
            const stage = runtime.stageOverviewPhaseStats(job)[stageKey] || {};
            return {
                handled: Number(stage.handled_count || 0),
                total: Number(stage.total_count || 0),
                success: Number(stage.success_count || 0),
                failed: Number(stage.failed_count || 0),
            };
        };

        runtime.stageOverviewStageProgressText = function (job, stageKey) {
            const stage = runtime.stageOverviewStageStat(job, stageKey);
            const pageStyle = !!(runtime.stageOverviewConfig && runtime.stageOverviewConfig.pageStylePhases && runtime.stageOverviewConfig.pageStylePhases[stageKey]);
            if (pageStyle) {
                if (stage.total <= 0) {
                    return '-';
                }
                return '第 ' + stage.handled + '/' + stage.total + ' 页';
            }
            if (stage.total <= 0) {
                return '-';
            }
            return String(stage.handled) + '/' + String(stage.total);
        };

        runtime.stageOverviewStageResultText = function (job, stageKey) {
            const stage = runtime.stageOverviewStageStat(job, stageKey);
            return String(stage.success) + ' / ' + String(stage.failed);
        };

        runtime.renderStageOverview = function (job) {
            const phaseKeys = runtime.stageOverviewConfig && Array.isArray(runtime.stageOverviewConfig.phaseKeys)
                ? runtime.stageOverviewConfig.phaseKeys
                : [];
            const activeStage = runtime.stageOverviewCurrentStageKey(job);
            if (runtime.currentStageText) {
                const labels = runtime.stageOverviewConfig && runtime.stageOverviewConfig.labels ? runtime.stageOverviewConfig.labels : {};
                runtime.currentStageText.textContent = activeStage ? (labels[activeStage] || activeStage) : '-';
            }
            if (runtime.elapsedText) {
                const elapsed = job ? runtime.formatElapsed(job.started_at) : '00:00';
                runtime.elapsedText.textContent = '已运行时长 ' + elapsed;
            }
            const slot1 = phaseKeys[0] || '';
            const slot2 = phaseKeys[1] || '';
            const slot3 = phaseKeys[2] || '';
            const slot4 = phaseKeys[3] || '';
            if (runtime.dailyBestBestinvProgressText) {
                runtime.dailyBestBestinvProgressText.textContent = slot1 ? runtime.stageOverviewStageProgressText(job, slot1) : '-';
            }
            if (runtime.dailyBestBestinvResultText) {
                runtime.dailyBestBestinvResultText.textContent = slot1 ? runtime.stageOverviewStageResultText(job, slot1) : '0 / 0';
            }
            if (runtime.dailyBestDetailProgressText) {
                runtime.dailyBestDetailProgressText.textContent = slot2 ? runtime.stageOverviewStageProgressText(job, slot2) : '-';
            }
            if (runtime.dailyBestDetailResultText) {
                runtime.dailyBestDetailResultText.textContent = slot2 ? runtime.stageOverviewStageResultText(job, slot2) : '0 / 0';
            }
            if (runtime.dailyBestCoverProgressText) {
                runtime.dailyBestCoverProgressText.textContent = slot3 ? runtime.stageOverviewStageProgressText(job, slot3) : '-';
            }
            if (runtime.dailyBestCoverResultText) {
                runtime.dailyBestCoverResultText.textContent = slot3 ? runtime.stageOverviewStageResultText(job, slot3) : '0 / 0';
            }
            if (runtime.dailyBestTranslateProgressText) {
                runtime.dailyBestTranslateProgressText.textContent = slot4 ? runtime.stageOverviewStageProgressText(job, slot4) : '-';
            }
            if (runtime.dailyBestTranslateResultText) {
                runtime.dailyBestTranslateResultText.textContent = slot4 ? runtime.stageOverviewStageResultText(job, slot4) : '0 / 0';
            }
            if (runtime.dailyBestFetchJavbusProgressText) {
                runtime.dailyBestFetchJavbusProgressText.textContent = runtime.stageOverviewStageProgressText(job, 'fetch_javbus');
            }
            if (runtime.dailyBestFetchJavbusResultText) {
                runtime.dailyBestFetchJavbusResultText.textContent = runtime.stageOverviewStageResultText(job, 'fetch_javbus');
            }
            if (runtime.dailyBestFetchSukebeiProgressText) {
                runtime.dailyBestFetchSukebeiProgressText.textContent = runtime.stageOverviewStageProgressText(job, 'fetch_sukebei');
            }
            if (runtime.dailyBestFetchSukebeiResultText) {
                runtime.dailyBestFetchSukebeiResultText.textContent = runtime.stageOverviewStageResultText(job, 'fetch_sukebei');
            }

            runtime.dailyBestStageCards.forEach(function (card) {
                const phase = card.getAttribute('data-phase') || '';
                card.classList.toggle('is-active', phase !== '' && phase === activeStage);
            });
        };

        runtime.currentJob = function () {
            const selectedId = runtime.normalizeJobId(runtime.state.selectedJobId);
            if (runtime.state.selectedJob && (!selectedId || runtime.state.selectedJob.job_id === selectedId)) {
                return runtime.state.selectedJob;
            }
            if (!selectedId) {
                return null;
            }
            const list = runtime.filterVisibleJobs(runtime.state.jobs);
            const matched = list.find(function (item) {
                return item && item.job_id === selectedId;
            }) || null;
            if (matched) {
                runtime.state.selectedJob = matched;
            }
            return matched;
        };

        runtime.isSelectedJobMode = function () {
            return runtime.normalizeJobId(runtime.state.selectedJobId) !== '';
        };

        runtime.syncSelectedJobIdToUrl = function (jobId) {
            if (!global.history || typeof global.history.replaceState !== 'function' || !global.location) {
                return;
            }
            try {
                const url = new URL(global.location.href);
                const clean = runtime.normalizeJobId(jobId);
                if (clean) {
                    url.searchParams.set('job_id', clean);
                } else {
                    url.searchParams.delete('job_id');
                }
                global.history.replaceState({}, '', url.toString());
            } catch (error) {
                return;
            }
        };

        runtime.showMessage = function (text, ok) {
            if (!runtime.msgArea) {
                return;
            }
            runtime.msgArea.textContent = String(text || '');
            runtime.msgArea.className = 'alert ' + (ok === false ? 'alert-danger' : 'alert-success') + ' fade show small py-2 px-3';
            runtime.msgArea.style.display = 'block';
            global.clearTimeout(runtime.msgTimer);
            runtime.msgTimer = global.setTimeout(function () {
                runtime.msgArea.style.display = 'none';
            }, 2600);
        };

        runtime.request = function (url, options) {
            return fetch(url, options).then(async function (response) {
                const data = await response.json().catch(function () {
                    return {};
                });
                if (!response.ok) {
                    throw new Error(data.error || ('请求失败(' + response.status + ')'));
                }
                return data;
            });
        };

        runtime.toggleField = function (element, visible) {
            if (!element) {
                return;
            }
            element.classList.toggle('is-hidden', !visible);
        };

        runtime.isFetchSiteTaskType = function (taskType) {
            const current = String(taskType || '').trim();
            return FETCH_SITE_TASK_TYPES.indexOf(current) >= 0 || current === TASK_FETCH_SITE_BOTH || current === TASK_FETCH_SUKEBEI_FILTERED || current === TASK_FETCH_JAVBUS_FILTERED;
        };

        runtime.fetchSiteStatsFromJob = function (job) {
            if (!job) {
                return {job: null, handled: 0, queued: 0, total: 0, success: 0, failed: 0};
            }
            const handled = Number(job.handled_count || 0);
            const queued = Number(job.queued_count || 0);
            return {
                job: job,
                handled: handled,
                queued: queued,
                total: handled + queued,
                success: Number(job.success_count || 0),
                failed: Number(job.failed_count || 0),
            };
        };

        runtime.isAutoFetchSiteTaskType = function (taskType) {
            const current = String(taskType || '').trim();
            return current === 'spider_daily_best' ||
                current === 'spider_daily_best_sync' ||
                current === 'spider_seeds' ||
                current === 'spider_seed_by_name';
        };

        runtime.isFetchSehuatangTaskType = function (taskType) {
            return String(taskType || '').trim() === TASK_FETCH_SEHUATANG;
        };

        runtime.isFetchSehuatangSHTaskType = function (taskType) {
            return String(taskType || '').trim() === TASK_FETCH_SEHUATANG_SH;
        };

        runtime.isFetchSehuatangSHListTaskType = function (taskType) {
            return String(taskType || '').trim() === TASK_FETCH_SEHUATANG_SH_LIST;
        };

        runtime.isFetchSehuatangSHDetailTaskType = function (taskType) {
            return String(taskType || '').trim() === TASK_FETCH_SEHUATANG_SH_DETAIL;
        };

        runtime.checkedSehuatangSHForumIDs = function () {
            return Array.from(runtime.form.querySelectorAll('input[name="target_forum_ids"]:checked')).map(function (input) {
                return String(input && input.value || '').trim();
            }).filter(function (value) {
                return value !== '';
            });
        };

        runtime.checkedSHDetailStatuses = function () {
            return Array.from(runtime.form.querySelectorAll('input[name="detail_statuses"]:checked')).map(function (input) {
                return String(input && input.value || '').trim();
            }).filter(function (value) {
                return value !== '';
            });
        };

        runtime.fetchSehuatangSHMode = function () {
            return String(runtime.fetchSehuatangSHModeInput && runtime.fetchSehuatangSHModeInput.value || 'search').trim() === 'forum' ? 'forum' : 'search';
        };

        runtime.syncFetchSehuatangStopMode = function () {
            if (!runtime.fetchSehuatangStopOnExistingInput || !runtime.fetchSehuatangStartPageInput) {
                return;
            }
            const stopOnExisting = !!runtime.fetchSehuatangStopOnExistingInput.checked;
            runtime.fetchSehuatangStartPageInput.disabled = stopOnExisting;
            if (stopOnExisting) {
                runtime.fetchSehuatangStartPageInput.value = '1';
            }
        };

        runtime.syncFetchSiteFilterToggleText = function () {
            if (!runtime.toggleFetchSiteFilterBtn) {
                return;
            }
            runtime.toggleFetchSiteFilterBtn.textContent = runtime.state.fetchSiteFilterExpanded ? '收起筛选' : '展开筛选';
        };

        runtime.syncTaskFields = function () {
            const taskType = String(runtime.taskTypeInput && runtime.taskTypeInput.value || '');
            const isFetchSiteTask = runtime.isFetchSiteTaskType(taskType);
            const isFetchSehuatangTask = runtime.isFetchSehuatangTaskType(taskType);
            const isFetchSehuatangSHTask = runtime.isFetchSehuatangSHTaskType(taskType);
            const isFetchSehuatangSHListTask = runtime.isFetchSehuatangSHListTaskType(taskType);
            const isFetchSehuatangSHDetailTask = runtime.isFetchSehuatangSHDetailTaskType(taskType);
            const isFetchSehuatangSHSearch = isFetchSehuatangSHTask && runtime.fetchSehuatangSHMode() === 'search';
            runtime.toggleField(runtime.nameFieldWrap, taskType === 'spider_seed_by_name');
            runtime.toggleField(
                runtime.numberFieldWrap,
                taskType === 'spider_refresh_oldest_detail' ||
                taskType === 'spider_fetch_javbus_resources' ||
                taskType === 'spider_fetch_sukebei_resources' ||
                taskType === TASK_FETCH_SITE_BOTH
            );
            runtime.toggleField(runtime.actorNameFieldWrap, taskType === 'spider_rebuild_actor_rank');
            runtime.toggleField(
                runtime.translateStatusesFieldWrap,
                taskType === 'spider_translate_title' || taskType === 'spider_post_process'
            );
            runtime.toggleField(runtime.autoFetchSiteFieldWrap, runtime.isAutoFetchSiteTaskType(taskType));
            runtime.toggleField(runtime.fetchSehuatangListURLFieldWrap, isFetchSehuatangTask);
            runtime.toggleField(runtime.fetchSehuatangKeywordFieldWrap, isFetchSehuatangTask || isFetchSehuatangSHSearch);
            runtime.toggleField(runtime.fetchSehuatangStopOnExistingFieldWrap, isFetchSehuatangTask);
            runtime.toggleField(runtime.fetchSehuatangStartPageFieldWrap, isFetchSehuatangTask || isFetchSehuatangSHTask || isFetchSehuatangSHListTask);
            runtime.toggleField(runtime.fetchSehuatangEndPageFieldWrap, isFetchSehuatangTask || isFetchSehuatangSHTask || isFetchSehuatangSHListTask);
            runtime.toggleField(runtime.fetchSehuatangPersistModeFieldWrap, isFetchSehuatangTask);
            runtime.toggleField(runtime.fetchSehuatangSHModeFieldWrap, isFetchSehuatangSHTask);
            runtime.toggleField(runtime.fetchSehuatangSHForumsFieldWrap, isFetchSehuatangSHTask || isFetchSehuatangSHListTask || isFetchSehuatangSHDetailTask);
            runtime.toggleField(runtime.fetchSehuatangSHDetailStatusFieldWrap, isFetchSehuatangSHDetailTask);
            runtime.toggleField(runtime.fetchSehuatangSHDetailTitleFieldWrap, isFetchSehuatangSHDetailTask);
            runtime.toggleField(runtime.fetchSehuatangSHPostTimeFromFieldWrap, isFetchSehuatangSHDetailTask);
            runtime.toggleField(runtime.fetchSehuatangSHPostTimeToFieldWrap, isFetchSehuatangSHDetailTask);
            runtime.toggleField(runtime.fetchSehuatangSHDetailLimitFieldWrap, isFetchSehuatangSHDetailTask);
            runtime.syncFetchSehuatangStopMode();
            if (!isFetchSiteTask) {
                runtime.state.fetchSiteFilterExpanded = false;
                runtime.hideFetchSitePreview();
            }
            if (!isFetchSehuatangSHDetailTask) {
                runtime.hideSHDetailPreview();
            }
            runtime.toggleField(runtime.fetchSiteFilterPanel, isFetchSiteTask);
            runtime.toggleField(runtime.fetchSiteFilterWrap, isFetchSiteTask && runtime.state.fetchSiteFilterExpanded);
            runtime.syncFetchSiteFilterToggleText();
        };

        runtime.ensureFetchSiteDefaultValues = function () {
            if (runtime.numberInput && String(runtime.numberInput.value || '').trim() === '') {
                runtime.numberInput.value = String(FETCH_SITE_DEFAULT_NUMBER);
            }
            if (!runtime.form) {
                return;
            }
            FETCH_SITE_DURATION_FIELD_NAMES.forEach(function (name) {
                const field = runtime.form.querySelector('[name="' + name + '"]');
                if (!field) {
                    return;
                }
                if (String(field.value || '').trim() !== '') {
                    return;
                }
                field.value = String(FETCH_SITE_DEFAULT_DURATION_DAYS);
            });
        };

        runtime.setTaskType = function (taskType) {
            const next = String(taskType || '').trim();
            if (!next || !runtime.taskTypeInput) {
                return;
            }
            runtime.taskTypeInput.value = next;
            page.querySelectorAll('[data-task-type]').forEach(function (button) {
                button.classList.toggle('active', button.getAttribute('data-task-type') === next);
            });
            runtime.syncTaskFields();
            if (runtime.isFetchSiteTaskType(next)) {
                runtime.ensureFetchSiteDefaultValues();
            }
        };

        runtime.currentPayload = function () {
            const taskType = String(runtime.taskTypeInput && runtime.taskTypeInput.value || '').trim();
            const payload = {task_type: taskType};
            const requiredField = REQUIRED_FIELDS[taskType];

            if (taskType === 'spider_seed_by_name') {
                payload.name = String(runtime.nameInput && runtime.nameInput.value || '').trim();
            }
            if (
                taskType === 'spider_refresh_oldest_detail' ||
                taskType === 'spider_fetch_javbus_resources' ||
                taskType === 'spider_fetch_sukebei_resources' ||
                taskType === TASK_FETCH_SITE_BOTH
            ) {
                payload.number = String(runtime.numberInput && runtime.numberInput.value || '').trim();
            }
            if (taskType === 'spider_rebuild_actor_rank') {
                payload.actor_name = String(runtime.actorNameInput && runtime.actorNameInput.value || '').trim();
            }
            if (taskType === 'spider_translate_title' || taskType === 'spider_post_process') {
                const checkedStatuses = Array.from(runtime.form.querySelectorAll('input[name="statuses"]:checked')).map(function (input) {
                    return String(input && input.value || '').trim();
                }).filter(function (value) {
                    return value !== '';
                });
                if (checkedStatuses.length > 0) {
                    payload.statuses = checkedStatuses;
                }
            }
            if (runtime.isFetchSehuatangTaskType(taskType)) {
                payload.list_url = String(runtime.fetchSehuatangListURLInput && runtime.fetchSehuatangListURLInput.value || '').trim();
                payload.keyword = String(runtime.fetchSehuatangKeywordInput && runtime.fetchSehuatangKeywordInput.value || '').trim();
                payload.start_page = String(runtime.fetchSehuatangStartPageInput && runtime.fetchSehuatangStartPageInput.value || '').trim();
                payload.end_page = String(runtime.fetchSehuatangEndPageInput && runtime.fetchSehuatangEndPageInput.value || '').trim();
                payload.persist_mode = String(runtime.fetchSehuatangPersistModeInput && runtime.fetchSehuatangPersistModeInput.value || '').trim();
                payload.stop_on_existing = !!(runtime.fetchSehuatangStopOnExistingInput && runtime.fetchSehuatangStopOnExistingInput.checked);
                const defaults = runtime.config && runtime.config.fetchSehuatangDefaults ? runtime.config.fetchSehuatangDefaults : {};
                if (payload.list_url === '') {
                    payload.list_url = String(defaults.listUrl || '').trim();
                }
                if (payload.stop_on_existing) {
                    payload.start_page = '1';
                }
                if (payload.start_page === '') {
                    payload.start_page = String(defaults.startPage || '').trim();
                }
                if (payload.end_page === '') {
                    payload.end_page = String(defaults.endPage || '').trim();
                }
                if (payload.persist_mode === '') {
                    payload.persist_mode = String(defaults.persistMode || 'skip_existing').trim();
                }
                const startPageNumber = Number(payload.start_page);
                const endPageNumber = Number(payload.end_page);
                if (!Number.isFinite(startPageNumber) || startPageNumber <= 0) {
                    throw new Error('请填写合法的起始页');
                }
                if (!Number.isFinite(endPageNumber) || endPageNumber <= 0) {
                    throw new Error('请填写合法的结束页');
                }
                if (endPageNumber < startPageNumber) {
                    throw new Error('结束页不能小于起始页');
                }
                payload.start_page = startPageNumber;
                payload.end_page = endPageNumber;
                if (payload.persist_mode !== 'skip_existing' && payload.persist_mode !== 'upsert_all') {
                    payload.persist_mode = 'skip_existing';
                }
            }
            if (runtime.isFetchSehuatangSHTaskType(taskType)) {
                payload.fetch_mode = runtime.fetchSehuatangSHMode();
                payload.keyword = String(runtime.fetchSehuatangKeywordInput && runtime.fetchSehuatangKeywordInput.value || '').trim();
                payload.start_page = String(runtime.fetchSehuatangStartPageInput && runtime.fetchSehuatangStartPageInput.value || '').trim();
                payload.end_page = String(runtime.fetchSehuatangEndPageInput && runtime.fetchSehuatangEndPageInput.value || '').trim();
                payload.target_forum_ids = runtime.checkedSehuatangSHForumIDs();
                const defaults = runtime.config && runtime.config.fetchSehuatangDefaults ? runtime.config.fetchSehuatangDefaults : {};
                if (payload.start_page === '') {
                    payload.start_page = String(defaults.startPage || '').trim();
                }
                if (payload.end_page === '') {
                    payload.end_page = String(defaults.endPage || '').trim();
                }
                const startPageNumber = Number(payload.start_page);
                const endPageNumber = Number(payload.end_page);
                if (!Number.isFinite(startPageNumber) || startPageNumber <= 0) {
                    throw new Error('请填写合法的起始页');
                }
                if (!Number.isFinite(endPageNumber) || endPageNumber <= 0) {
                    throw new Error('请填写合法的结束页');
                }
                if (endPageNumber < startPageNumber) {
                    throw new Error('结束页不能小于起始页');
                }
                if (payload.fetch_mode === 'search' && payload.keyword === '') {
                    throw new Error('请填写搜索关键词');
                }
                if (!payload.target_forum_ids.length) {
                    throw new Error('请选择目标版块');
                }
                payload.start_page = startPageNumber;
                payload.end_page = endPageNumber;
            }
            if (runtime.isFetchSehuatangSHListTaskType(taskType)) {
                payload.start_page = String(runtime.fetchSehuatangStartPageInput && runtime.fetchSehuatangStartPageInput.value || '').trim();
                payload.end_page = String(runtime.fetchSehuatangEndPageInput && runtime.fetchSehuatangEndPageInput.value || '').trim();
                payload.target_forum_ids = runtime.checkedSehuatangSHForumIDs();
                const defaults = runtime.config && runtime.config.fetchSehuatangDefaults ? runtime.config.fetchSehuatangDefaults : {};
                if (payload.start_page === '') {
                    payload.start_page = String(defaults.startPage || '').trim();
                }
                if (payload.end_page === '') {
                    payload.end_page = String(defaults.endPage || '').trim();
                }
                const startPageNumber = Number(payload.start_page);
                const endPageNumber = Number(payload.end_page);
                if (!Number.isFinite(startPageNumber) || startPageNumber <= 0) {
                    throw new Error('请填写合法的起始页');
                }
                if (!Number.isFinite(endPageNumber) || endPageNumber <= 0) {
                    throw new Error('请填写合法的结束页');
                }
                if (endPageNumber < startPageNumber) {
                    throw new Error('结束页不能小于起始页');
                }
                if (!payload.target_forum_ids.length) {
                    throw new Error('请选择目标版块');
                }
                payload.start_page = startPageNumber;
                payload.end_page = endPageNumber;
            }
            if (runtime.isFetchSehuatangSHDetailTaskType(taskType)) {
                payload.thread_tids = runtime.checkedSHDetailPreviewThreadTIDs();
                if (!payload.thread_tids.length) {
                    throw new Error('请先筛选帖子预览并选择要抓取详情的帖子');
                }
            }
            if (runtime.isAutoFetchSiteTaskType(taskType)) {
                payload.auto_fetch_site = runtime.autoFetchSiteInput && runtime.autoFetchSiteInput.checked ? '1' : '0';
            }
            if (runtime.isFetchSiteTaskType(taskType)) {
                runtime.ensureFetchSiteDefaultValues();
                payload.number = String(runtime.numberInput && runtime.numberInput.value || '').trim();
                runtime.appendMovieFilters(payload);
                runtime.appendFetchSiteDurationFilters(payload);
            }
            if (runtime.isFetchSiteTaskType(taskType) && String(payload.number || '').trim() === '') {
                payload.number = String(FETCH_SITE_DEFAULT_NUMBER);
            }

            if (requiredField && !String(payload[requiredField] || '').trim()) {
                throw new Error('请填写任务参数');
            }
            if (
                taskType === 'spider_refresh_oldest_detail' ||
                taskType === 'spider_fetch_javbus_resources' ||
                taskType === 'spider_fetch_sukebei_resources' ||
                taskType === TASK_FETCH_SITE_BOTH
            ) {
                const parsed = Number(payload.number);
                if (!Number.isFinite(parsed)) {
                    throw new Error('请填写合法的数量');
                }
                payload.number = parsed;
            }
            return payload;
        };

        runtime.appendMovieFilters = function (payload) {
            MOVIE_FILTER_FIELD_NAMES.forEach(function (name) {
                const field = runtime.form ? runtime.form.querySelector('[name="' + name + '"]') : null;
                if (!field) {
                    return;
                }
                const value = String(field.value || '').trim();
                if (value === '') {
                    return;
                }
                payload[name] = value;
            });
        };

        runtime.clearMovieFilters = function () {
            if (!runtime.form) {
                return;
            }
            MOVIE_FILTER_FIELD_NAMES.forEach(function (name) {
                const field = runtime.form.querySelector('[name="' + name + '"]');
                if (!field) {
                    return;
                }
                if (name === 'od') {
                    field.value = 'rd';
                    return;
                }
                if (name === 'order') {
                    field.value = '';
                    return;
                }
                field.value = '';
            });
            FETCH_SITE_DURATION_FIELD_NAMES.forEach(function (name) {
                const field = runtime.form.querySelector('[name="' + name + '"]');
                if (!field) {
                    return;
                }
                field.value = String(FETCH_SITE_DEFAULT_DURATION_DAYS);
            });
            if (runtime.numberInput && String(runtime.numberInput.value || '').trim() === '') {
                runtime.numberInput.value = String(FETCH_SITE_DEFAULT_NUMBER);
            }
            runtime.syncFetchSiteOrderByButtons();
            runtime.hideFetchSitePreview();
        };

        runtime.hideFetchSitePreview = function () {
            runtime.state.fetchSitePreviewPayload = null;
            runtime.state.fetchSitePreviewPage = 1;
            runtime.state.fetchSitePreviewTotalPages = 0;
            if (runtime.fetchSitePreviewPanel) {
                runtime.fetchSitePreviewPanel.classList.add('is-hidden');
            }
            if (runtime.fetchSitePreviewSummary) {
                runtime.fetchSitePreviewSummary.textContent = '当前还没有预览数据';
            }
            if (runtime.fetchSitePreviewMetaText) {
                runtime.fetchSitePreviewMetaText.textContent = '共 0 条';
            }
            if (runtime.fetchSitePreviewPageText) {
                runtime.fetchSitePreviewPageText.textContent = '第 0/0 页';
            }
            if (runtime.fetchSitePreviewPrevBtn) {
                runtime.fetchSitePreviewPrevBtn.disabled = true;
            }
            if (runtime.fetchSitePreviewNextBtn) {
                runtime.fetchSitePreviewNextBtn.disabled = true;
            }
            if (runtime.fetchSitePreviewTableBody) {
                runtime.fetchSitePreviewTableBody.innerHTML = '<tr><td colspan="4" class="text-center text-muted py-4">等待预览</td></tr>';
            }
        };

        runtime.renderFetchSitePreview = function (payload) {
            const result = payload || {};
            const items = Array.isArray(result.items) ? result.items : [];
            const total = Number(result.total || 0);
            const pageNumber = Number(result.page || 1);
            const pageSize = Number(result.page_size || runtime.state.fetchSitePreviewPageSize || 100);
            const totalPages = Number(result.total_pages || 0);
            runtime.state.fetchSitePreviewPage = pageNumber;
            runtime.state.fetchSitePreviewTotalPages = totalPages;
            runtime.state.fetchSitePreviewPageSize = pageSize;

            if (runtime.fetchSitePreviewPanel) {
                runtime.fetchSitePreviewPanel.classList.remove('is-hidden');
            }
            if (runtime.fetchSitePreviewSummary) {
                runtime.fetchSitePreviewSummary.textContent = runtime.taskLabel(result.task_type) + ' 预览结果';
            }
            if (runtime.fetchSitePreviewMetaText) {
                runtime.fetchSitePreviewMetaText.textContent = '共 ' + String(total) + ' 条';
            }
            if (runtime.fetchSitePreviewPageText) {
                runtime.fetchSitePreviewPageText.textContent = '第 ' + String(pageNumber) + '/' + String(totalPages) + ' 页';
            }
            if (runtime.fetchSitePreviewPrevBtn) {
                runtime.fetchSitePreviewPrevBtn.disabled = pageNumber <= 1;
            }
            if (runtime.fetchSitePreviewNextBtn) {
                runtime.fetchSitePreviewNextBtn.disabled = totalPages <= 0 || pageNumber >= totalPages;
            }
            if (!runtime.fetchSitePreviewTableBody) {
                return;
            }
            if (!items.length) {
                runtime.fetchSitePreviewTableBody.innerHTML = '<tr><td colspan="4" class="text-center text-muted py-4">当前筛选没有命中影片</td></tr>';
                return;
            }
            runtime.fetchSitePreviewTableBody.innerHTML = items.map(function (item, index) {
                const seq = (pageNumber - 1) * pageSize + index + 1;
                const movieName = runtime.escapeHtml(String(item && item.movie_name || '').trim() || '-');
                const movieJavID = runtime.escapeHtml(String(item && item.movie_jav_id || '').trim() || '-');
                const source = runtime.escapeHtml(String(item && item.source || '').trim() || '-');
                return '' +
                    '<tr>' +
                    '<td class="text-nowrap">' + String(seq) + '</td>' +
                    '<td class="text-nowrap">' + movieName + '</td>' +
                    '<td class="text-nowrap">' + movieJavID + '</td>' +
                    '<td class="text-nowrap">' + source + '</td>' +
                    '</tr>';
            }).join('');
        };

        runtime.previewFetchSite = function (pageNumber, useStoredPayload) {
            const taskType = String(runtime.taskTypeInput && runtime.taskTypeInput.value || '').trim();
            if (!runtime.isFetchSiteTaskType(taskType)) {
                runtime.showMessage('当前任务不支持筛选预览', false);
                return;
            }
            let payload;
            if (useStoredPayload && runtime.state.fetchSitePreviewPayload) {
                payload = Object.assign({}, runtime.state.fetchSitePreviewPayload);
            } else {
                payload = runtime.currentPayload();
                runtime.state.fetchSitePreviewPayload = Object.assign({}, payload);
            }
            payload.page = pageNumber > 0 ? pageNumber : 1;
            payload.page_size = runtime.state.fetchSitePreviewPageSize || 100;
            if (runtime.previewFetchSiteBtn) {
                runtime.previewFetchSiteBtn.disabled = true;
            }
            runtime.request(API_BASE + '/fetch-site/preview', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(payload),
            }).then(function (data) {
                runtime.renderFetchSitePreview(data || {});
                if (runtime.previewFetchSiteBtn) {
                    runtime.previewFetchSiteBtn.disabled = false;
                }
            }).catch(function (error) {
                if (runtime.previewFetchSiteBtn) {
                    runtime.previewFetchSiteBtn.disabled = false;
                }
                runtime.showMessage(error.message, false);
            });
        };

        runtime.formatSHUnixTime = function (value) {
            const ts = Number(value || 0);
            if (!Number.isFinite(ts) || ts <= 0) {
                return '-';
            }
            const date = new Date(ts * 1000);
            const pad = function (num) {
                return String(num).padStart(2, '0');
            };
            return [
                date.getFullYear(),
                pad(date.getMonth() + 1),
                pad(date.getDate()),
            ].join('-') + ' ' + [
                pad(date.getHours()),
                pad(date.getMinutes()),
                pad(date.getSeconds()),
            ].join(':');
        };

        runtime.currentSHDetailFilterPayload = function () {
            const payload = {
                task_type: TASK_FETCH_SEHUATANG_SH_DETAIL,
                target_forum_ids: runtime.checkedSehuatangSHForumIDs(),
                detail_statuses: runtime.checkedSHDetailStatuses(),
                keyword: String(runtime.fetchSehuatangSHDetailKeywordInput && runtime.fetchSehuatangSHDetailKeywordInput.value || '').trim(),
                post_time_from: String(runtime.fetchSehuatangSHPostTimeFromInput && runtime.fetchSehuatangSHPostTimeFromInput.value || '').trim(),
                post_time_to: String(runtime.fetchSehuatangSHPostTimeToInput && runtime.fetchSehuatangSHPostTimeToInput.value || '').trim(),
            };
            const limitText = String(runtime.fetchSehuatangSHDetailLimitInput && runtime.fetchSehuatangSHDetailLimitInput.value || '').trim();
            if (limitText !== '') {
                const limitValue = Number(limitText);
                if (!Number.isFinite(limitValue) || limitValue <= 0) {
                    throw new Error('请填写合法的 limit');
                }
                payload.limit = Math.floor(limitValue);
            }
            if (!payload.target_forum_ids.length) {
                throw new Error('请选择目标版块');
            }
            return payload;
        };

        runtime.checkedSHDetailPreviewThreadTIDs = function () {
            if (!runtime.shDetailPreviewTableBody) {
                return [];
            }
            return Array.from(runtime.shDetailPreviewTableBody.querySelectorAll('input[name="sh_detail_thread_tid"]:checked')).map(function (input) {
                return String(input && input.value || '').trim();
            }).filter(function (value) {
                return value !== '';
            });
        };

        runtime.updateSHDetailPreviewSelectionText = function () {
            const total = Array.isArray(runtime.state.shDetailPreviewRows) ? runtime.state.shDetailPreviewRows.length : 0;
            const selected = runtime.checkedSHDetailPreviewThreadTIDs().length;
            if (runtime.shDetailPreviewMetaText) {
                runtime.shDetailPreviewMetaText.textContent = '共 ' + String(total) + ' 条，已选 ' + String(selected) + ' 条';
            }
        };

        runtime.hideSHDetailPreview = function () {
            runtime.state.shDetailPreviewRows = [];
            if (runtime.shDetailPreviewPanel) {
                runtime.shDetailPreviewPanel.classList.add('is-hidden');
            }
            if (runtime.shDetailPreviewSummary) {
                runtime.shDetailPreviewSummary.textContent = '当前还没有预览数据';
            }
            if (runtime.shDetailPreviewMetaText) {
                runtime.shDetailPreviewMetaText.textContent = '共 0 条，已选 0 条';
            }
            if (runtime.shDetailPreviewTableBody) {
                runtime.shDetailPreviewTableBody.innerHTML = '<tr><td colspan="7" class="text-center text-muted py-4">等待预览</td></tr>';
            }
        };

        runtime.renderSHDetailPreview = function (payload) {
            const rows = Array.isArray(payload && payload.rows) ? payload.rows : [];
            const total = Number(payload && payload.total || 0);
            runtime.state.shDetailPreviewRows = rows;
            if (runtime.shDetailPreviewPanel) {
                runtime.shDetailPreviewPanel.classList.remove('is-hidden');
            }
            if (runtime.shDetailPreviewSummary) {
                runtime.shDetailPreviewSummary.textContent = '筛选命中 ' + String(total) + ' 条，当前预览 ' + String(rows.length) + ' 条';
            }
            if (!runtime.shDetailPreviewTableBody) {
                runtime.updateSHDetailPreviewSelectionText();
                return;
            }
            if (!rows.length) {
                runtime.shDetailPreviewTableBody.innerHTML = '<tr><td colspan="7" class="text-center text-muted py-4">当前筛选没有命中帖子</td></tr>';
                runtime.updateSHDetailPreviewSelectionText();
                return;
            }
            runtime.shDetailPreviewTableBody.innerHTML = rows.map(function (row, index) {
                const threadTID = String((row && (row.ThreadTID || row.thread_tid)) || '').trim();
                const title = runtime.escapeHtml(String((row && (row.ThreadTitle || row.thread_title)) || '').trim() || '-');
                const forumID = runtime.escapeHtml(String((row && (row.SourceForumID || row.source_forum_id)) || '').trim() || '-');
                const postTime = runtime.escapeHtml(runtime.formatSHUnixTime(row && (row.PostTime || row.post_time)));
                const detailStatus = runtime.escapeHtml(String((row && (row.DetailStatus || row.detail_status)) || '').trim() || '-');
                const authorName = runtime.escapeHtml(String((row && (row.AuthorName || row.author_name)) || '').trim() || '-');
                const inputID = 'sh_detail_preview_' + String(index);
                return '' +
                    '<tr>' +
                    '<td class="text-nowrap"><input id="' + inputID + '" class="form-check-input sh-detail-preview-check" type="checkbox" name="sh_detail_thread_tid" value="' + runtime.escapeHtml(threadTID) + '" checked></td>' +
                    '<td class="text-nowrap">' + runtime.escapeHtml(threadTID || '-') + '</td>' +
                    '<td><label class="mb-0" for="' + inputID + '">' + title + '</label></td>' +
                    '<td class="text-nowrap">' + forumID + '</td>' +
                    '<td class="text-nowrap">' + postTime + '</td>' +
                    '<td class="text-nowrap">' + detailStatus + '</td>' +
                    '<td class="text-nowrap">' + authorName + '</td>' +
                    '</tr>';
            }).join('');
            Array.from(runtime.shDetailPreviewTableBody.querySelectorAll('input[name="sh_detail_thread_tid"]')).forEach(function (input) {
                input.addEventListener('change', function () {
                    runtime.updateSHDetailPreviewSelectionText();
                });
            });
            runtime.updateSHDetailPreviewSelectionText();
        };

        runtime.previewSHDetail = function () {
            let payload;
            try {
                payload = runtime.currentSHDetailFilterPayload();
            } catch (error) {
                runtime.showMessage(error.message, false);
                return;
            }
            if (runtime.previewSHDetailBtn) {
                runtime.previewSHDetailBtn.disabled = true;
            }
            runtime.request(API_BASE + '/sehuatang-sh-thread-detail/preview', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(payload),
            }).then(function (data) {
                runtime.renderSHDetailPreview(data || {});
                if (runtime.previewSHDetailBtn) {
                    runtime.previewSHDetailBtn.disabled = false;
                }
            }).catch(function (error) {
                if (runtime.previewSHDetailBtn) {
                    runtime.previewSHDetailBtn.disabled = false;
                }
                runtime.showMessage(error.message, false);
            });
        };

        runtime.syncFetchSiteOrderByButtons = function () {
            if (!runtime.form) {
                return;
            }
            const orderByField = runtime.form.querySelector('[name="od"]');
            const current = String(orderByField && orderByField.value || '').trim();
            page.querySelectorAll('[data-fetch-site-orderby]').forEach(function (button) {
                const active = String(button.getAttribute('data-fetch-site-orderby') || '').trim() === current;
                button.classList.toggle('active', active);
            });
        };

        runtime.setFetchSiteOrderBy = function (value) {
            if (!runtime.form) {
                return;
            }
            const orderByField = runtime.form.querySelector('[name="od"]');
            if (!orderByField) {
                return;
            }
            orderByField.value = String(value || '').trim() || 'rd';
            runtime.syncFetchSiteOrderByButtons();
        };

        runtime.appendFetchSiteDurationFilters = function (payload) {
            FETCH_SITE_DURATION_FIELD_NAMES.forEach(function (name) {
                const field = runtime.form ? runtime.form.querySelector('[name="' + name + '"]') : null;
                if (!field) {
                    return;
                }
                const value = String(field.value || '').trim();
                if (value === '') {
                    return;
                }
                payload[name] = value;
            });
        };

        runtime.closeStream = function () {
            if (runtime.state.eventSource) {
                runtime.state.eventSource.close();
                runtime.state.eventSource = null;
            }
            runtime.state.eventStreamKey = '';
            if (runtime.state.reconnectTimer) {
                global.clearTimeout(runtime.state.reconnectTimer);
                runtime.state.reconnectTimer = null;
            }
        };

        runtime.clearTimers = function () {
            if (runtime.state.loadTimer) {
                global.clearInterval(runtime.state.loadTimer);
                runtime.state.loadTimer = null;
            }
            if (runtime.state.elapsedTimer) {
                global.clearInterval(runtime.state.elapsedTimer);
                runtime.state.elapsedTimer = null;
            }
        };

        runtime.destroy = function () {
            if (runtime.state.destroyed) {
                return;
            }
            runtime.state.destroyed = true;
            runtime.closeStream();
            runtime.clearTimers();
        };

        runtime.connectEventStream = function () {
            if (runtime.state.destroyed) {
                return;
            }

            const selectedId = runtime.normalizeJobId(runtime.state.selectedJobId);
            const streamKey = selectedId ? ('job:' + selectedId) : 'group';
            const streamURL = selectedId
                ? (API_BASE + '/' + encodeURIComponent(selectedId) + '/stream')
                : (API_BASE + '/stream');
            const lastEventId = Number(runtime.state.lastEventId || 0);
            const streamWithCursor = lastEventId > 0
                ? (streamURL + '?last_event_id=' + encodeURIComponent(String(lastEventId)))
                : streamURL;

            if (runtime.state.eventSource && runtime.state.eventStreamKey === streamKey) {
                return;
            }

            runtime.closeStream();

            const source = new EventSource(streamWithCursor);
            runtime.state.eventSource = source;
            runtime.state.eventStreamKey = streamKey;
            source.onmessage = function (message) {
                if (runtime.state.eventSource !== source) {
                    return;
                }
                try {
                    runtime.handleIncomingEvent(JSON.parse(message.data));
                } catch (error) {
                    runtime.showMessage('事件解析失败', false);
                }
            };
            source.onerror = function () {
                if (runtime.state.destroyed || runtime.state.eventSource !== source) {
                    return;
                }
                const currentJob = runtime.currentJob();
                if (selectedId && currentJob && currentJob.done) {
                    runtime.closeStream();
                    return;
                }
                source.close();
                if (runtime.state.eventSource === source) {
                    runtime.state.eventSource = null;
                    runtime.state.eventStreamKey = '';
                }
                if (!runtime.state.reconnectTimer) {
                    runtime.state.reconnectTimer = global.setTimeout(function () {
                        runtime.state.reconnectTimer = null;
                        runtime.connectEventStream();
                    }, 1500);
                }
            };
        };

        runtime.updateControls = function () {
            const currentTaskType = String(runtime.taskTypeInput && runtime.taskTypeInput.value || '').trim();
            if (currentTaskType === TASK_FETCH_SITE_BOTH) {
                const pauseJobIDs = runtime.resolveFetchSiteBothControlJobIDs('pause');
                const resumeJobIDs = runtime.resolveFetchSiteBothControlJobIDs('resume');
                const stopJobIDs = runtime.resolveFetchSiteBothControlJobIDs('stop');
                if (runtime.pauseBtn) runtime.pauseBtn.disabled = pauseJobIDs.length <= 0;
                if (runtime.resumeBtn) runtime.resumeBtn.disabled = resumeJobIDs.length <= 0;
                if (runtime.stopBtn) runtime.stopBtn.disabled = stopJobIDs.length <= 0;
                return;
            }
            const job = runtime.currentJob();
            const selectedId = runtime.normalizeJobId(runtime.state.selectedJobId);
            const paused = !!(job && job.paused);
            const done = !!(job && job.done);
            if (runtime.pauseBtn) runtime.pauseBtn.disabled = !job || paused || done;
            if (runtime.resumeBtn) runtime.resumeBtn.disabled = !job || !paused || done;
            if (runtime.stopBtn) runtime.stopBtn.disabled = (!job && !selectedId) || done;
        };

        runtime.jobSortTs = function (job) {
            return Number((job && (job.at || job.started_at)) || 0);
        };

        runtime.pickLatestJobByTaskTypeWithMatcher = function (taskType, matcher) {
            const list = runtime.filterVisibleJobs(runtime.state.jobs).filter(function (item) {
                if (!item || String(item.task_type || '') !== taskType) {
                    return false;
                }
                if (typeof matcher !== 'function') {
                    return true;
                }
                return !!matcher(item);
            });
            if (!list.length) {
                return null;
            }
            list.sort(function (a, b) {
                return runtime.jobSortTs(b) - runtime.jobSortTs(a);
            });
            return list[0] || null;
        };

        runtime.isJobActionAllowed = function (job, action) {
            if (!job) {
                return action !== 'resume';
            }
            const done = !!job.done;
            const paused = !!job.paused;
            switch (action) {
                case 'pause':
                    return !done && !paused;
                case 'resume':
                    return !done && paused;
                case 'stop':
                    return !done;
                default:
                    return false;
            }
        };

        runtime.resolveFetchSiteBothControlJobIDs = function (action) {
            const visibleJobs = runtime.filterVisibleJobs(runtime.state.jobs);
            const jobByID = {};
            visibleJobs.forEach(function (item) {
                const id = runtime.normalizeJobId(item && item.job_id);
                if (id) {
                    jobByID[id] = item;
                }
            });
            const picked = [];
            const pickedSet = {};
            const addPicked = function (jobID) {
                const clean = runtime.normalizeJobId(jobID);
                if (!clean || pickedSet[clean]) {
                    return;
                }
                pickedSet[clean] = true;
                picked.push(clean);
            };

            const remembered = Array.isArray(runtime.state.fetchSiteBothJobIds) ? runtime.state.fetchSiteBothJobIds : [];
            remembered.forEach(function (jobID) {
                const clean = runtime.normalizeJobId(jobID);
                if (!clean) {
                    return;
                }
                const job = jobByID[clean];
                if (runtime.isJobActionAllowed(job, action)) {
                    addPicked(clean);
                }
            });

            FETCH_SITE_TASK_TYPES.forEach(function (taskType) {
                const latest = runtime.pickLatestJobByTaskTypeWithMatcher(taskType, function (item) {
                    return runtime.isJobActionAllowed(item, action);
                });
                if (latest) {
                    addPicked(latest.job_id);
                }
            });
            return picked;
        };

        runtime.pickLatestJobByTaskType = function (taskType) {
            const list = runtime.filterVisibleJobs(runtime.state.jobs).filter(function (item) {
                return item && String(item.task_type || '') === taskType;
            });
            if (!list.length) {
                return null;
            }
            let picked = list[0];
            let pickedTs = Number((picked && (picked.at || picked.started_at)) || 0);
            for (let i = 1; i < list.length; i += 1) {
                const current = list[i];
                const currentTs = Number((current && (current.at || current.started_at)) || 0);
                if (currentTs > pickedTs) {
                    picked = current;
                    pickedTs = currentTs;
                }
            }
            return picked || null;
        };

        runtime.fetchSiteStatsForTask = function (taskType) {
            const job = runtime.pickLatestJobByTaskType(taskType);
            return runtime.fetchSiteStatsFromJob(job);
        };

        runtime.fetchSiteSummaryState = function (javbusStats, sukebeiStats) {
            const jobs = [javbusStats.job, sukebeiStats.job].filter(function (item) {
                return !!item;
            });
            if (!jobs.length) {
                return {className: 'text-muted', text: runtime.emptyStateText};
            }
            const hasFailed = jobs.some(function (item) {
                return !!(item.done && item.stage === 'failed');
            });
            if (hasFailed) {
                return {className: 'err', text: '存在失败任务'};
            }
            const hasRunning = jobs.some(function (item) {
                return !item.done;
            });
            if (hasRunning) {
                return {className: 'text-muted', text: '任务运行中'};
            }
            return {className: 'ok', text: '任务完成'};
        };

        runtime.queueBadgeClass = function (state) {
            switch (String(state || '').trim()) {
                case 'running':
                    return 'is-running';
                case 'success':
                    return 'is-success';
                case 'failed':
                    return 'is-failed';
                default:
                    return 'is-pending';
            }
        };

        runtime.queueBadgeText = function (state) {
            switch (String(state || '').trim()) {
                case 'running':
                    return '处理中';
                case 'success':
                    return '成功';
                case 'failed':
                    return '失败';
                default:
                    return '待入列';
            }
        };

        runtime.renderQueueItems = function (element, items, emptyText, reverseOrder) {
            if (!element) {
                return;
            }
            const safeItems = runtime.normalizeQueueItems(items);
            const renderItems = reverseOrder ? safeItems.slice().reverse() : safeItems;
            if (!safeItems.length) {
                element.innerHTML = '<div class="queue-empty">' + runtime.escapeHtml(emptyText) + '</div>';
                return;
            }
            element.innerHTML = renderItems.map(function (item, index) {
                const movieLabel = item.movie_name || '-';
                const movieHref = item.movie_name ? ('/movie/' + encodeURIComponent(item.movie_name)) : '';
                const indexText = String(item.seq > 0 ? item.seq : (index + 1)).padStart(2, '0');
                const titleHtml = movieHref
                    ? '<a class="queue-item-name" href="' + movieHref + '">' + runtime.escapeHtml(movieLabel) + '</a>'
                    : '<div class="queue-item-name">' + runtime.escapeHtml(movieLabel) + '</div>';
                const completeHtml = item.completed_at > 0
                    ? '<div class="queue-item-meta">完成时间：' + runtime.escapeHtml(runtime.timeText(item.completed_at)) + '</div>'
                    : '';
                const errorHtml = item.error
                    ? '<div class="queue-item-error">' + runtime.escapeHtml(item.error) + '</div>'
                    : '';
                return '' +
                    '<div class="queue-item">' +
                    '<div class="queue-item-top">' +
                    '<div class="queue-item-left"><span class="queue-item-index">' + runtime.escapeHtml(indexText) + '</span>' + titleHtml + '</div>' +
                    '<span class="queue-badge ' + runtime.queueBadgeClass(item.state) + '">' + runtime.escapeHtml(runtime.queueBadgeText(item.state)) + '</span>' +
                    '</div>' +
                    completeHtml +
                    errorHtml +
                    '</div>';
            }).join('');
        };

        runtime.renderOverview = function (payload) {
            const job = runtime.currentJob();
            if (runtime.stageOverviewConfig) {
                runtime.renderStageOverview(job);
                if (!job) {
                    const selectedId = runtime.normalizeJobId(runtime.state.selectedJobId);
                    if (runtime.progressBar) {
                        runtime.progressBar.style.width = '0%';
                        runtime.progressBar.textContent = '0%';
                        runtime.progressBar.setAttribute('aria-valuenow', '0');
                    }
                    if (runtime.statusMsg) {
                        runtime.statusMsg.className = 'text-muted';
                        runtime.statusMsg.textContent = selectedId ? '正在获取任务状态' : runtime.emptyStateText;
                    }
                    runtime.updateControls();
                    return;
                }
                if (runtime.progressBar) {
                    const percent = runtime.progressPercent(job);
                    runtime.progressBar.style.width = percent + '%';
                    runtime.progressBar.textContent = percent + '%';
                    runtime.progressBar.setAttribute('aria-valuenow', String(percent));
                }
                if (runtime.statusMsg) {
                    runtime.statusMsg.className = job.done && job.stage === 'failed' ? 'err' : (job.done ? 'ok' : 'text-muted');
                    runtime.statusMsg.textContent = (job.message || runtime.stageLabel(job.stage || '')) || runtime.emptyStateText;
                }
                runtime.updateControls();
                return;
            }
            if (runtime.overviewExtraMode === 'fetch_site_summary') {
                const javbusStats = runtime.fetchSiteStatsForTask('spider_fetch_javbus_resources');
                const sukebeiStats = runtime.fetchSiteStatsForTask('spider_fetch_sukebei_resources');
                if (!javbusStats.job && !sukebeiStats.job) {
                    const selectedId = runtime.normalizeJobId(runtime.state.selectedJobId);
                    if (runtime.fetchSiteJavbusProgressText) runtime.fetchSiteJavbusProgressText.textContent = '0 / 0';
                    if (runtime.fetchSiteJavbusResultText) runtime.fetchSiteJavbusResultText.textContent = '0 / 0';
                    if (runtime.fetchSiteSukebeiProgressText) runtime.fetchSiteSukebeiProgressText.textContent = '0 / 0';
                    if (runtime.fetchSiteSukebeiResultText) runtime.fetchSiteSukebeiResultText.textContent = '0 / 0';
                    if (runtime.progressBar) {
                        runtime.progressBar.style.width = '0%';
                        runtime.progressBar.textContent = '0%';
                        runtime.progressBar.setAttribute('aria-valuenow', '0');
                    }
                    if (runtime.statusMsg) {
                        runtime.statusMsg.className = 'text-muted';
                        runtime.statusMsg.textContent = selectedId ? '正在获取任务状态' : runtime.emptyStateText;
                    }
                    runtime.updateControls();
                    return;
                }

                if (runtime.fetchSiteJavbusProgressText) runtime.fetchSiteJavbusProgressText.textContent = String(javbusStats.total) + ' / ' + String(javbusStats.handled);
                if (runtime.fetchSiteJavbusResultText) runtime.fetchSiteJavbusResultText.textContent = String(javbusStats.success) + ' / ' + String(javbusStats.failed);
                if (runtime.fetchSiteSukebeiProgressText) runtime.fetchSiteSukebeiProgressText.textContent = String(sukebeiStats.total) + ' / ' + String(sukebeiStats.handled);
                if (runtime.fetchSiteSukebeiResultText) runtime.fetchSiteSukebeiResultText.textContent = String(sukebeiStats.success) + ' / ' + String(sukebeiStats.failed);
                if (runtime.progressBar) {
                    const total = javbusStats.total + sukebeiStats.total;
                    const handled = javbusStats.handled + sukebeiStats.handled;
                    const percent = total > 0 ? Math.max(0, Math.min(100, Math.round(handled * 100 / total))) : 0;
                    runtime.progressBar.style.width = percent + '%';
                    runtime.progressBar.textContent = percent + '%';
                    runtime.progressBar.setAttribute('aria-valuenow', String(percent));
                }
                if (runtime.statusMsg) {
                    const summaryState = runtime.fetchSiteSummaryState(javbusStats, sukebeiStats);
                    runtime.statusMsg.className = summaryState.className;
                    runtime.statusMsg.textContent = summaryState.text;
                }
                runtime.updateControls();
                return;
            }
            if (runtime.overviewExtraMode === 'sukebei_filtered_queue' || runtime.overviewExtraMode === 'javbus_filtered_queue') {
                const pendingItems = runtime.normalizeQueueItems(job && job.pending_items);
                const runningItems = runtime.normalizeQueueItems(job && job.running_items);
                const doneItems = runtime.normalizeQueueItems(job && job.done_items);
                const total = pendingItems.length + runningItems.length + doneItems.length;
                const successCount = Number((job || {}).success_count || 0);
                const failedCount = Number((job || {}).failed_count || 0);
                const selectedId = runtime.normalizeJobId(runtime.state.selectedJobId);

                if (runtime.filteredQueueTotalText) runtime.filteredQueueTotalText.textContent = String(total);
                if (runtime.filteredQueuePendingText) runtime.filteredQueuePendingText.textContent = String(pendingItems.length);
                if (runtime.filteredQueueSuccessText) runtime.filteredQueueSuccessText.textContent = String(successCount);
                if (runtime.filteredQueueFailedText) runtime.filteredQueueFailedText.textContent = String(failedCount);

                runtime.renderQueueItems(runtime.filteredQueuePendingList, pendingItems, selectedId ? '当前任务没有待入列影片' : '等待任务启动', false);
                runtime.renderQueueItems(runtime.filteredQueueDoneList, doneItems, selectedId ? '当前还没有处理完成的影片' : '等待任务启动', true);

                if (!job) {
                    if (runtime.progressBar) {
                        runtime.progressBar.style.width = '0%';
                        runtime.progressBar.textContent = '0%';
                        runtime.progressBar.setAttribute('aria-valuenow', '0');
                    }
                    if (runtime.statusMsg) {
                        runtime.statusMsg.className = 'text-muted';
                        runtime.statusMsg.textContent = selectedId ? '正在获取任务状态' : runtime.emptyStateText;
                    }
                    if (runtime.elapsedText) {
                        runtime.elapsedText.textContent = '00:00';
                    }
                    runtime.updateControls();
                    return;
                }

                if (runtime.progressBar) {
                    const percent = total > 0 ? Math.max(0, Math.min(100, Math.round(doneItems.length * 100 / total))) : 0;
                    runtime.progressBar.style.width = percent + '%';
                    runtime.progressBar.textContent = percent + '%';
                    runtime.progressBar.setAttribute('aria-valuenow', String(percent));
                }
                if (runtime.statusMsg) {
                    runtime.statusMsg.className = job.done && job.stage === 'failed' ? 'err' : (job.done ? 'ok' : 'text-muted');
                    runtime.statusMsg.textContent = (job.message || runtime.stageLabel(job.stage || '')) || runtime.emptyStateText;
                }
                if (runtime.elapsedText) {
                    runtime.elapsedText.textContent = runtime.formatElapsed(job.started_at);
                }
                runtime.updateControls();
                return;
            }
            if (runtime.overviewExtraMode === 'sehuatang_progress') {
                const setSehuatangText = function (element, value) {
                    if (element) {
                        element.textContent = String(value);
                    }
                };
                if (!job) {
                    const selectedId = runtime.normalizeJobId(runtime.state.selectedJobId);
                    setSehuatangText(runtime.sehuatangProgressText, '0 / 0');
                    setSehuatangText(runtime.sehuatangPageProgressText, '0 / 0');
                    setSehuatangText(runtime.sehuatangCurrentPageText, '-');
                    setSehuatangText(runtime.sehuatangSeenText, '0');
                    setSehuatangText(runtime.sehuatangMatchedText, '0');
                    setSehuatangText(runtime.sehuatangSkippedForumText, '0');
                    setSehuatangText(runtime.sehuatangSkippedExistingText, '0');
                    setSehuatangText(runtime.sehuatangFetchedText, '0');
                    setSehuatangText(runtime.sehuatangFailedText, '0');
                    if (runtime.progressBar) {
                        runtime.progressBar.style.width = '0%';
                        runtime.progressBar.textContent = '0%';
                        runtime.progressBar.setAttribute('aria-valuenow', '0');
                    }
                    if (runtime.statusMsg) {
                        runtime.statusMsg.className = 'text-muted';
                        runtime.statusMsg.textContent = selectedId ? '正在获取任务状态' : runtime.emptyStateText;
                    }
                    if (runtime.elapsedText) runtime.elapsedText.textContent = '00:00';
                    runtime.updateControls();
                    return;
                }

                const listPhaseStat = runtime.stageOverviewStageStat(job, 'list');
                let completedPageCount = runtime.jobStatNumber(job, 'completed_page_count');
                let totalPageCount = runtime.jobStatNumber(job, 'total_page_count');
                if (totalPageCount <= 0 && listPhaseStat.total > 0) {
                    completedPageCount = listPhaseStat.handled;
                    totalPageCount = listPhaseStat.total;
                }
                const seenCount = runtime.jobStatNumber(job, 'seen_count');
                const matchedCount = runtime.hasJobStat(job, 'matched_count') ? runtime.jobStatNumber(job, 'matched_count') : Number(job.handled_count || 0);
                const skippedForumCount = runtime.jobStatNumber(job, 'skipped_forum_count');
                const skippedExistingCount = runtime.jobStatNumber(job, 'skipped_existing_count');
                const fetchedCount = runtime.hasJobStat(job, 'fetched_count') ? runtime.jobStatNumber(job, 'fetched_count') : Number(job.success_count || 0);
                const failedCount = runtime.hasJobStat(job, 'failed_count') ? runtime.jobStatNumber(job, 'failed_count') : Number(job.failed_count || 0);
                let currentPageText = runtime.jobStatText(job, 'current_page_text');
                if (!currentPageText) {
                    const currentPage = runtime.jobStatNumber(job, 'current_page');
                    currentPageText = currentPage > 0 ? '第 ' + String(currentPage) + ' 页' : '-';
                }

                setSehuatangText(runtime.sehuatangProgressText, String(completedPageCount) + ' / ' + String(totalPageCount));
                setSehuatangText(runtime.sehuatangPageProgressText, String(completedPageCount) + ' / ' + String(totalPageCount));
                setSehuatangText(runtime.sehuatangCurrentPageText, currentPageText);
                setSehuatangText(runtime.sehuatangSeenText, seenCount);
                setSehuatangText(runtime.sehuatangMatchedText, matchedCount);
                setSehuatangText(runtime.sehuatangSkippedForumText, skippedForumCount);
                setSehuatangText(runtime.sehuatangSkippedExistingText, skippedExistingCount);
                setSehuatangText(runtime.sehuatangFetchedText, fetchedCount);
                setSehuatangText(runtime.sehuatangFailedText, failedCount);

                if (runtime.progressBar) {
                    const percent = totalPageCount > 0 ? Math.max(0, Math.min(100, Math.round(completedPageCount * 100 / totalPageCount))) : 0;
                    runtime.progressBar.style.width = percent + '%';
                    runtime.progressBar.textContent = percent + '%';
                    runtime.progressBar.setAttribute('aria-valuenow', String(percent));
                }
                if (runtime.statusMsg) {
                    runtime.statusMsg.className = job.done && job.stage === 'failed' ? 'err' : (job.done ? 'ok' : 'text-muted');
                    runtime.statusMsg.textContent = (job.message || runtime.stageLabel(job.stage || '')) || runtime.emptyStateText;
                }
                if (runtime.elapsedText) {
                    runtime.elapsedText.textContent = runtime.formatElapsed(job.started_at);
                }
                runtime.updateControls();
                return;
            }
            const detailLoop = payload && payload.detail_loop ? payload.detail_loop : runtime.state.detailLoop;

            if (runtime.detailLoopText) {
                if (runtime.overviewExtraMode === 'task_type') {
                    runtime.detailLoopText.textContent = job ? runtime.taskLabel(job.task_type) : '-';
                } else if (runtime.overviewExtraMode === 'page_progress') {
                    runtime.detailLoopText.textContent = job ? runtime.pageProgressText(job) : '-';
                } else if (!detailLoop) {
                    runtime.detailLoopText.textContent = '未知';
                } else if (detailLoop.running) {
                    runtime.detailLoopText.textContent = detailLoop.paused ? '已暂停' : '运行中';
                } else {
                    runtime.detailLoopText.textContent = '未启动';
                }
            }

            if (!job) {
                const selectedId = runtime.normalizeJobId(runtime.state.selectedJobId);
                if (runtime.resultText) runtime.resultText.textContent = '0 / 0';
                if (runtime.progressBar) {
                    runtime.progressBar.style.width = '0%';
                    runtime.progressBar.textContent = '0%';
                    runtime.progressBar.setAttribute('aria-valuenow', '0');
                }
                if (runtime.statusMsg) {
                    runtime.statusMsg.className = 'text-muted';
                    runtime.statusMsg.textContent = selectedId ? '正在获取任务状态' : runtime.emptyStateText;
                }
                if (runtime.elapsedText) runtime.elapsedText.textContent = '00:00';
                runtime.updateControls();
                return;
            }

            if (runtime.resultText) runtime.resultText.textContent = String(job.success_count || 0) + ' / ' + String(job.failed_count || 0);
            if (runtime.progressBar) {
                const percent = runtime.progressPercent(job);
                runtime.progressBar.style.width = percent + '%';
                runtime.progressBar.textContent = percent + '%';
                runtime.progressBar.setAttribute('aria-valuenow', String(percent));
            }
            if (runtime.statusMsg) {
                runtime.statusMsg.className = job.done && job.stage === 'failed' ? 'err' : (job.done ? 'ok' : 'text-muted');
                runtime.statusMsg.textContent = (job.message || runtime.stageLabel(job.stage || '')) || runtime.emptyStateText;
            }
            if (runtime.elapsedText) {
                runtime.elapsedText.textContent = runtime.formatElapsed(job.started_at);
            }
            runtime.updateControls();
        };

        runtime.nowTimeText = function (ts) {
            const source = ts ? new Date(ts * 1000) : new Date();
            const hh = String(source.getHours()).padStart(2, '0');
            const mm = String(source.getMinutes()).padStart(2, '0');
            const ss = String(source.getSeconds()).padStart(2, '0');
            return hh + ':' + mm + ':' + ss;
        };

        runtime.humanMessage = function (event) {
            const stage = String((event || {}).stage || '');
            switch (stage) {
                case 'job_started':
                    return '任务已启动';
                case 'paused':
                    return '任务已暂停';
                case 'resumed':
                    return '任务已继续';
                case 'done':
                    return '任务完成';
                case 'failed':
                    return '任务失败';
                default:
                    return String((event || {}).message || runtime.stageLabel(stage) || '收到事件');
            }
        };

        runtime.normalizeLogLevel = function (level) {
            const current = String(level || '').trim().toLowerCase();
            if (current === 'warning') {
                return 'warn';
            }
            if (current === 'error') {
                return 'error';
            }
            return current || 'info';
        };

        runtime.buildEventEntry = function (event) {
            if (!event) {
                return null;
            }
            const prefix = '[' + runtime.nowTimeText(event.at) + ']';
            if (runtime.debugRaw && runtime.debugRaw.checked) {
                return {
                    level: runtime.normalizeLogLevel(event.level),
                    text: prefix + ' ' + JSON.stringify(event, null, 2),
                };
            }
            if (event.kind === 'progress') {
                if (runtime.overviewExtraMode === 'fetch_site_summary') {
                    return null;
                }
                const stage = String(event.stage || '').trim();
                const message = String(event.message || runtime.stageLabel(stage) || '收到进度').trim();
                const parts = [prefix + ' ' + message];
                if (Number(event.handled_count || 0) || Number(event.queued_count || 0)) {
                    parts.push('handled=' + String(event.handled_count || 0));
                    parts.push('queued=' + String(event.queued_count || 0));
                }
                if (Number(event.success_count || 0) || Number(event.failed_count || 0)) {
                    parts.push('success=' + String(event.success_count || 0));
                    parts.push('failed=' + String(event.failed_count || 0));
                }
                return {
                    level: stage === 'failed' ? 'error' : (stage === 'paused' ? 'warn' : 'info'),
                    text: parts.join(' | '),
                };
            }
            if (event.kind !== 'log') {
                return null;
            }
            const rawLine = String(event.line || '').trim();
            const match = rawLine.match(/^\[([^\]]+)\]\s+(INFO|WARN|WARNING|ERROR)\s+(.*)$/i);
            if (match) {
                return {
                    level: runtime.normalizeLogLevel(match[2]),
                    text: '[' + match[1] + '] ' + match[3],
                };
            }
            return {
                level: runtime.normalizeLogLevel(event.level),
                text: prefix + ' ' + (rawLine || String(event.message || '')),
            };
        };

        runtime.formatEventLineHTML = function (entry) {
            const raw = String((entry && entry.text) || '');
            const parts = raw.split(' | ');
            if (parts.length < 5 || !/^\[[^\]]+\]$/.test(parts[0] || '')) {
                return runtime.escapeHtml(raw).replace(/\b([A-Za-z][A-Za-z0-9]*-\d{2,})\b/g, function (match, code) {
                    const href = '/movie/' + encodeURIComponent(code);
                    return '<a href="' + href + '" class="event-link">' + runtime.escapeHtml(code) + '</a>';
                });
            }
            const subjectId = String(parts[2] || '').trim();
            if (!/^\d+$/.test(subjectId) || String(parts[3] || '').indexOf('new_items=') !== 0) {
                return runtime.escapeHtml(raw).replace(/\b([A-Za-z][A-Za-z0-9]*-\d{2,})\b/g, function (match, code) {
                    const href = '/movie/' + encodeURIComponent(code);
                    return '<a href="' + href + '" class="event-link">' + runtime.escapeHtml(code) + '</a>';
                });
            }
            const prefix = runtime.escapeHtml(parts[0]) + ' | ';
            const movieName = runtime.escapeHtml(parts[1] || '');
            const suffix = parts.slice(3).map(function (part) {
                return runtime.escapeHtml(part);
            }).join(' | ');
            const localHref = '/movie/' + encodeURIComponent(subjectId);
            const doubanHref = 'https://movie.douban.com/subject/' + encodeURIComponent(subjectId) + '/';
            const movieLink = '<a href="' + localHref + '" target="_blank" rel="noopener noreferrer" class="event-link">' + movieName + '</a>';
            const subjectLink = '<a href="' + doubanHref + '" target="_blank" rel="noopener noreferrer">' + runtime.escapeHtml(subjectId) + '</a>';
            return prefix + movieLink + ' | ' + subjectLink + ' | ' + suffix;
        };

        runtime.renderEvents = function () {
            if (!runtime.resultBox) {
                return;
            }
            if (!runtime.state.eventEntries.length) {
                runtime.resultBox.textContent = runtime.emptyStateText;
                if (runtime.eventCountText) {
                    runtime.eventCountText.textContent = '0 条';
                }
                return;
            }
            runtime.resultBox.innerHTML = runtime.state.eventEntries.map(function (entry) {
                const level = runtime.normalizeLogLevel(entry && entry.level);
                return '<span class="event-line log-' + runtime.escapeHtml(level) + '">' + runtime.formatEventLineHTML(entry) + '</span>';
            }).join('');
            runtime.resultBox.scrollTop = 0;
            if (runtime.eventCountText) {
                runtime.eventCountText.textContent = String(runtime.state.eventCount) + ' 条';
            }
        };

        runtime.clearEvents = function () {
            runtime.state.eventEntries = [];
            runtime.state.eventCount = 0;
            runtime.state.lastEventId = 0;
            runtime.renderEvents();
        };

        runtime.appendEvent = function (event) {
            if (!event) {
                return;
            }
            const entry = runtime.buildEventEntry(event);
            if (!entry) {
                return;
            }
            runtime.state.eventEntries.unshift(entry);
            if (runtime.state.eventEntries.length > EVENT_LIMIT) {
                runtime.state.eventEntries.length = EVENT_LIMIT;
            }
            runtime.state.eventCount = Number(runtime.state.eventCount || 0) + 1;
            runtime.state.lastEventId = Math.max(Number(runtime.state.lastEventId || 0), Number(event.id || 0));
            runtime.renderEvents();
        };

        runtime.mergeSelectedJob = function (event) {
            if (!event || event.kind !== 'progress' || !event.job_id || !runtime.isVisibleTaskType(event.task_type)) {
                return;
            }
            if (runtime.state.selectedJobId !== event.job_id) {
                return;
            }
            const current = runtime.state.selectedJob || {};
            runtime.state.selectedJob = {
                job_id: runtime.pickValue(event.job_id, runtime.pickValue(current.job_id, runtime.state.selectedJobId)),
                task_type: runtime.pickValue(event.task_type, current.task_type),
                stage: runtime.pickValue(event.stage, current.stage),
                message: runtime.pickValue(event.message, current.message),
                handled_count: runtime.pickNumber(event.handled_count, current.handled_count),
                success_count: runtime.pickNumber(event.success_count, current.success_count),
                failed_count: runtime.pickNumber(event.failed_count, current.failed_count),
                queued_count: runtime.pickNumber(event.queued_count, current.queued_count),
                pending_items: event.pending_items !== undefined ? runtime.normalizeQueueItems(event.pending_items) : runtime.normalizeQueueItems(current.pending_items),
                running_items: event.running_items !== undefined ? runtime.normalizeQueueItems(event.running_items) : runtime.normalizeQueueItems(current.running_items),
                done_items: event.done_items !== undefined ? runtime.normalizeQueueItems(event.done_items) : runtime.normalizeQueueItems(current.done_items),
                current_phase_key: String(runtime.pickValue(event.current_phase_key, current.current_phase_key) || ''),
                phase_stats: runtime.pickValue(event.phase_stats, current.phase_stats) || null,
                stats: runtime.pickValue(event.stats, current.stats) || null,
                stats_text: runtime.pickValue(event.stats_text, current.stats_text) || null,
                started_at: runtime.pickNumber(event.started_at, current.started_at),
                at: runtime.pickNumber(event.at, current.at),
                done: !!event.done,
                paused: event.stage === 'paused' ? true : (event.stage === 'resumed' ? false : !!current.paused),
            };
        };

        runtime.mergeJobList = function (event) {
            if (!event || event.kind !== 'progress' || !event.job_id || !runtime.isVisibleTaskType(event.task_type)) {
                return;
            }
            const currentList = Array.isArray(runtime.state.jobs) ? runtime.state.jobs.slice() : [];
            const index = currentList.findIndex(function (item) {
                return item && item.job_id === event.job_id;
            });
            const current = index >= 0 ? (currentList[index] || {}) : {};
            const next = {
                job_id: runtime.pickValue(event.job_id, runtime.pickValue(current.job_id, '')),
                task_type: runtime.pickValue(event.task_type, runtime.pickValue(current.task_type, '')),
                stage: runtime.pickValue(event.stage, runtime.pickValue(current.stage, '')),
                message: runtime.pickValue(event.message, runtime.pickValue(current.message, '')),
                handled_count: runtime.pickNumber(event.handled_count, current.handled_count),
                success_count: runtime.pickNumber(event.success_count, current.success_count),
                failed_count: runtime.pickNumber(event.failed_count, current.failed_count),
                queued_count: runtime.pickNumber(event.queued_count, current.queued_count),
                pending_items: event.pending_items !== undefined ? runtime.normalizeQueueItems(event.pending_items) : runtime.normalizeQueueItems(current.pending_items),
                running_items: event.running_items !== undefined ? runtime.normalizeQueueItems(event.running_items) : runtime.normalizeQueueItems(current.running_items),
                done_items: event.done_items !== undefined ? runtime.normalizeQueueItems(event.done_items) : runtime.normalizeQueueItems(current.done_items),
                current_phase_key: String(runtime.pickValue(event.current_phase_key, current.current_phase_key) || ''),
                phase_stats: runtime.pickValue(event.phase_stats, current.phase_stats) || null,
                stats: runtime.pickValue(event.stats, current.stats) || null,
                stats_text: runtime.pickValue(event.stats_text, current.stats_text) || null,
                started_at: runtime.pickNumber(event.started_at, current.started_at),
                at: runtime.pickNumber(event.at, current.at),
                done: !!event.done,
                paused: event.stage === 'paused' ? true : (event.stage === 'resumed' ? false : !!current.paused),
            };
            if (index >= 0) {
                currentList[index] = next;
            } else {
                currentList.push(next);
            }
            runtime.state.jobs = currentList;
        };

        runtime.loadJobs = function (silent) {
            runtime.request(API_BASE)
                .then(function (payload) {
                    runtime.state.jobs = Array.isArray(payload.jobs) ? payload.jobs : [];
                    runtime.state.detailLoop = payload.detail_loop || null;

                    const visibleJobs = runtime.filterVisibleJobs(runtime.state.jobs);
                    const selectedId = runtime.normalizeJobId(runtime.state.selectedJobId);
                    let selected = null;
                    if (selectedId) {
                        selected = visibleJobs.find(function (item) {
                            return item && item.job_id === selectedId;
                        }) || null;
                        if (selected) {
                            runtime.state.selectedJob = selected;
                        } else if (runtime.state.selectedJob && runtime.state.selectedJob.job_id === selectedId) {
                            selected = runtime.state.selectedJob;
                        } else {
                            runtime.state.selectedJob = null;
                        }
                    } else {
                        runtime.state.selectedJob = null;
                    }

                    runtime.renderOverview(payload);
                })
                .catch(function (error) {
                    if (!silent) {
                        runtime.showMessage(error.message, false);
                    }
                });
        };

        runtime.fetchJob = function (jobId) {
            return runtime.request(API_BASE + '/' + encodeURIComponent(jobId));
        };

        runtime.loadJobEvents = function (jobId) {
            const clean = runtime.normalizeJobId(jobId);
            if (!clean) {
                runtime.clearEvents();
                return Promise.resolve();
            }

            return runtime.request(API_BASE + '/' + encodeURIComponent(clean) + '/events')
                .then(function (payload) {
                    if (runtime.state.selectedJobId !== clean) {
                        return;
                    }
                    const events = Array.isArray((payload || {}).events) ? payload.events : [];
                    const lines = [];
                    events.forEach(function (event) {
                        if (!event || event.job_id !== clean || !runtime.isVisibleTaskType(event.task_type)) {
                            return;
                        }
                        const entry = runtime.buildEventEntry(event);
                        if (!entry) {
                            return;
                        }
                        lines.unshift(entry);
                        runtime.state.lastEventId = Math.max(Number(runtime.state.lastEventId || 0), Number(event.id || 0));
                    });
                    runtime.state.eventEntries = lines.slice(0, EVENT_LIMIT);
                    runtime.state.eventCount = lines.length;
                    runtime.renderEvents();
                    runtime.renderOverview();
                })
                .catch(function (error) {
                    if (runtime.state.selectedJobId === clean) {
                        runtime.showMessage(error.message, false);
                    }
                });
        };

        runtime.handleIncomingEvent = function (event) {
            if (!event || !runtime.isVisibleTaskType(event.task_type)) {
                return;
            }
            runtime.mergeJobList(event);
            runtime.mergeSelectedJob(event);
            const selectedId = runtime.normalizeJobId(runtime.state.selectedJobId);
            if ((selectedId && event.job_id === selectedId) || (!selectedId && runtime.isVisibleTaskType(event.task_type))) {
                runtime.appendEvent(event);
            }
            runtime.renderOverview();
        };

        runtime.openJob = function (jobId) {
            const clean = runtime.normalizeJobId(jobId);
            if (!clean) {
                runtime.state.selectedJobId = '';
                runtime.state.selectedJob = null;
                global.localStorage.removeItem(runtime.storageKey);
                runtime.clearEvents();
                runtime.syncSelectedJobIdToUrl('');
                runtime.renderOverview();
                runtime.connectEventStream();
                return;
            }

            runtime.state.selectedJobId = clean;
            global.localStorage.setItem(runtime.storageKey, clean);
            runtime.syncSelectedJobIdToUrl(clean);
            runtime.clearEvents();

            const fallbackJob = runtime.filterVisibleJobs(runtime.state.jobs).find(function (item) {
                return item && item.job_id === clean;
            }) || null;
            runtime.state.selectedJob = fallbackJob;
            runtime.renderOverview();
            runtime.loadJobEvents(clean).finally(function () {
                if (runtime.normalizeJobId(runtime.state.selectedJobId) === clean) {
                    runtime.connectEventStream();
                }
            });

            runtime.fetchJob(clean)
                .then(function (job) {
                    if (!runtime.isVisibleTaskType(job.task_type)) {
                        runtime.state.selectedJobId = '';
                        runtime.state.selectedJob = null;
                        global.localStorage.removeItem(runtime.storageKey);
                        runtime.clearEvents();
                        runtime.syncSelectedJobIdToUrl('');
                        runtime.renderOverview();
                        runtime.connectEventStream();
                        return;
                    }
                    runtime.state.selectedJob = job;
                    runtime.renderOverview();
                })
                .catch(function (error) {
                    if (fallbackJob) {
                        runtime.state.selectedJob = fallbackJob;
                        runtime.renderOverview();
                        return;
                    }
                    runtime.state.selectedJob = null;
                    runtime.renderOverview();
                    runtime.showMessage(error.message, false);
                });
        };

        runtime.startTask = function () {
            let payload;
            try {
                payload = runtime.currentPayload();
            } catch (error) {
                runtime.showMessage(error.message, false);
                return;
            }
            if (payload.task_type === TASK_FETCH_SITE_BOTH) {
                const startPayloads = FETCH_SITE_TASK_TYPES.map(function (taskType) {
                    const next = Object.assign({}, payload);
                    next.task_type = taskType;
                    return next;
                });
                runtime.state.fetchSiteBothJobIds = [];
                Promise.allSettled(startPayloads.map(function (item) {
                    return runtime.request(API_BASE + '/start', {
                        method: 'POST',
                        headers: {'Content-Type': 'application/json'},
                        body: JSON.stringify(item),
                    });
                })).then(function (results) {
                    const successCount = results.filter(function (item) {
                        return item.status === 'fulfilled';
                    }).length;
                    const startedJobIDs = results.map(function (item) {
                        if (item.status !== 'fulfilled') {
                            return '';
                        }
                        return runtime.normalizeJobId(item.value && item.value.job_id);
                    }).filter(function (item) {
                        return item !== '';
                    });
                    runtime.state.fetchSiteBothJobIds = startedJobIDs;
                    if (successCount <= 0) {
                        const firstError = results.find(function (item) {
                            return item.status === 'rejected';
                        });
                        const errorMessage = firstError && firstError.reason && firstError.reason.message
                            ? firstError.reason.message
                            : '任务启动失败';
                        runtime.showMessage(errorMessage, false);
                        return;
                    }
                    runtime.openJob('');
                    runtime.loadJobs(true);
                    if (successCount === startPayloads.length) {
                        runtime.showMessage('JavBus 与 Sukebei 已同时启动', true);
                        return;
                    }
                    const firstError = results.find(function (item) {
                        return item.status === 'rejected';
                    });
                    const errorMessage = firstError && firstError.reason && firstError.reason.message
                        ? firstError.reason.message
                        : '部分任务启动失败';
                    runtime.showMessage('部分启动成功（' + String(successCount) + '/' + String(startPayloads.length) + '）：' + errorMessage, false);
                });
                return;
            }
            runtime.request(API_BASE + '/start', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(payload),
            }).then(function (data) {
                runtime.showMessage(runtime.taskLabel(payload.task_type) + ' 已启动', true);
                runtime.openJob(data.job_id);
                runtime.loadJobs(true);
            }).catch(function (error) {
                runtime.showMessage(error.message, false);
            });
        };

        runtime.controlSelected = function (action) {
            const currentTaskType = String(runtime.taskTypeInput && runtime.taskTypeInput.value || '').trim();
            if (currentTaskType === TASK_FETCH_SITE_BOTH) {
                const actionLabelMap = {
                    pause: '暂停',
                    resume: '继续',
                    stop: '停止',
                };
                const actionLabel = actionLabelMap[action] || '操作';
                const targetJobIDs = runtime.resolveFetchSiteBothControlJobIDs(action);
                if (!targetJobIDs.length) {
                    runtime.showMessage('当前没有可' + actionLabel + '的同时抓取任务', false);
                    return;
                }
                Promise.allSettled(targetJobIDs.map(function (jobID) {
                    return runtime.request(API_BASE + '/' + encodeURIComponent(jobID) + '/' + action, {
                        method: 'POST',
                    });
                })).then(function (results) {
                    const successItems = results.filter(function (item) {
                        return item.status === 'fulfilled';
                    });
                    const failedItems = results.filter(function (item) {
                        return item.status === 'rejected';
                    });
                    if (!successItems.length) {
                        const firstError = failedItems[0];
                        const errorMessage = firstError && firstError.reason && firstError.reason.message
                            ? firstError.reason.message
                            : '操作失败';
                        runtime.showMessage(errorMessage, false);
                        return;
                    }
                    if (action === 'stop') {
                        runtime.state.fetchSiteBothJobIds = [];
                    }
                    if (!failedItems.length) {
                        runtime.showMessage('已' + actionLabel + '同时抓取任务（' + String(successItems.length) + '/' + String(results.length) + '）', true);
                        runtime.loadJobs(true);
                        return;
                    }
                    const firstError = failedItems[0];
                    const errorMessage = firstError && firstError.reason && firstError.reason.message
                        ? firstError.reason.message
                        : '部分操作失败';
                    runtime.showMessage('已' + actionLabel + ' ' + String(successItems.length) + ' 个任务，失败 ' + String(failedItems.length) + ' 个：' + errorMessage, false);
                    runtime.loadJobs(true);
                });
                return;
            }
            const job = runtime.currentJob();
            const jobID = runtime.normalizeJobId(runtime.state.selectedJobId || (job && job.job_id));
            if (!jobID) {
                runtime.showMessage('请先选择任务', false);
                return;
            }
            runtime.request(API_BASE + '/' + encodeURIComponent(jobID) + '/' + action, {
                method: 'POST',
            }).then(function (data) {
                runtime.showMessage(data.message || '操作完成', true);
                runtime.loadJobs(true);
            }).catch(function (error) {
                runtime.showMessage(error.message, false);
            });
        };

        runtime.bindEvents = function () {
            page.querySelectorAll('[data-task-type]').forEach(function (button) {
                button.addEventListener('click', function () {
                    runtime.setTaskType(button.getAttribute('data-task-type'));
                });
            });
            page.querySelectorAll('[data-fetch-site-orderby]').forEach(function (button) {
                button.addEventListener('click', function () {
                    runtime.setFetchSiteOrderBy(button.getAttribute('data-fetch-site-orderby'));
                });
            });

        if (runtime.form) {
            runtime.form.addEventListener('submit', function (event) {
                    event.preventDefault();
                    runtime.startTask();
            });
        }
        if (runtime.clearFetchSiteFilterBtn) {
            runtime.clearFetchSiteFilterBtn.addEventListener('click', function () {
                runtime.clearMovieFilters();
            });
        }
        if (runtime.previewFetchSiteBtn) {
            runtime.previewFetchSiteBtn.addEventListener('click', function () {
                runtime.previewFetchSite(1, false);
            });
        }
        if (runtime.previewSHDetailBtn) {
            runtime.previewSHDetailBtn.addEventListener('click', function () {
                runtime.previewSHDetail();
            });
        }
        if (runtime.selectAllSHDetailPreviewBtn) {
            runtime.selectAllSHDetailPreviewBtn.addEventListener('click', function () {
                if (!runtime.shDetailPreviewTableBody) {
                    return;
                }
                Array.from(runtime.shDetailPreviewTableBody.querySelectorAll('input[name="sh_detail_thread_tid"]')).forEach(function (input) {
                    input.checked = true;
                });
                runtime.updateSHDetailPreviewSelectionText();
            });
        }
        if (runtime.clearSHDetailPreviewBtn) {
            runtime.clearSHDetailPreviewBtn.addEventListener('click', function () {
                if (!runtime.shDetailPreviewTableBody) {
                    return;
                }
                Array.from(runtime.shDetailPreviewTableBody.querySelectorAll('input[name="sh_detail_thread_tid"]')).forEach(function (input) {
                    input.checked = false;
                });
                runtime.updateSHDetailPreviewSelectionText();
            });
        }
        if (runtime.hideSHDetailPreviewBtn) {
            runtime.hideSHDetailPreviewBtn.addEventListener('click', function () {
                runtime.hideSHDetailPreview();
            });
        }
        if (runtime.hideFetchSitePreviewBtn) {
            runtime.hideFetchSitePreviewBtn.addEventListener('click', function () {
                runtime.hideFetchSitePreview();
            });
        }
        if (runtime.fetchSitePreviewPrevBtn) {
            runtime.fetchSitePreviewPrevBtn.addEventListener('click', function () {
                if (runtime.state.fetchSitePreviewPage > 1) {
                    runtime.previewFetchSite(runtime.state.fetchSitePreviewPage - 1, true);
                }
            });
        }
        if (runtime.fetchSitePreviewNextBtn) {
            runtime.fetchSitePreviewNextBtn.addEventListener('click', function () {
                if (runtime.state.fetchSitePreviewTotalPages > 0 && runtime.state.fetchSitePreviewPage < runtime.state.fetchSitePreviewTotalPages) {
                    runtime.previewFetchSite(runtime.state.fetchSitePreviewPage + 1, true);
                }
            });
        }
        if (runtime.toggleFetchSiteFilterBtn) {
            runtime.toggleFetchSiteFilterBtn.addEventListener('click', function () {
                const taskType = String(runtime.taskTypeInput && runtime.taskTypeInput.value || '');
                if (!runtime.isFetchSiteTaskType(taskType)) {
                    return;
                }
                runtime.state.fetchSiteFilterExpanded = !runtime.state.fetchSiteFilterExpanded;
                runtime.syncTaskFields();
            });
        }
        if (runtime.fetchSehuatangStopOnExistingInput) {
            runtime.fetchSehuatangStopOnExistingInput.addEventListener('change', function () {
                runtime.syncFetchSehuatangStopMode();
            });
        }
        if (runtime.fetchSehuatangSHModeInput) {
            runtime.fetchSehuatangSHModeInput.addEventListener('change', function () {
                runtime.syncTaskFields();
            });
        }

            if (runtime.pauseBtn) {
                runtime.pauseBtn.addEventListener('click', function () {
                    runtime.controlSelected('pause');
                });
            }
            if (runtime.resumeBtn) {
                runtime.resumeBtn.addEventListener('click', function () {
                    runtime.controlSelected('resume');
                });
            }
            if (runtime.stopBtn) {
                runtime.stopBtn.addEventListener('click', function () {
                    runtime.controlSelected('stop');
                });
            }
            if (runtime.debugRaw) {
                runtime.debugRaw.addEventListener('change', function () {
                    runtime.renderEvents();
                });
            }
        };

        runtime.startTimers = function () {
            runtime.state.loadTimer = global.setInterval(function () {
                runtime.loadJobs(true);
            }, 5000);
            runtime.state.elapsedTimer = global.setInterval(function () {
                runtime.renderOverview();
            }, 1000);
        };

        runtime.bindLifecycle = function () {
            global.addEventListener('beforeunload', runtime.destroy);
            global.addEventListener('pagehide', runtime.destroy);
        };

        runtime.init = function () {
            const configuredJobId = runtime.normalizeJobId(runtime.config.initialJobId);
            const storedRawJobId = global.localStorage.getItem(runtime.storageKey);
            runtime.syncFetchSiteOrderByButtons();
            runtime.ensureFetchSiteDefaultValues();
            const storedJobId = runtime.normalizeJobId(storedRawJobId);
            const initialJobId = configuredJobId || storedJobId;

            if (storedRawJobId && !storedJobId) {
                global.localStorage.removeItem(runtime.storageKey);
            }

            runtime.bindEvents();
            runtime.bindLifecycle();
            runtime.setTaskType(runtime.defaultTaskType || String(runtime.taskTypeInput && runtime.taskTypeInput.value || 'spider_daily_best'));
            if (initialJobId) {
                runtime.state.selectedJobId = initialJobId;
                runtime.openJob(initialJobId);
            } else {
                runtime.clearEvents();
                runtime.connectEventStream();
            }
            runtime.loadJobs(false);
            runtime.startTimers();
        };

        runtime.init();
        return runtime;
    }

    global.createCrawlerJobConsole = createCrawlerJobConsole;

    if (document.getElementById('crawlerJobsPage')) {
        createCrawlerJobConsole(global.CRAWLER_JOBS_PAGE_DATA || {});
    }
})(window);
