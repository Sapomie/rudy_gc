(function (global) {
    const API_BASE = '/api/crawler/jobs';
    const TASK_LABELS = {
        spider_daily_best: 'DailyBest 抓取',
        spider_daily_best_sync: 'DailyBest 同步',
        spider_seeds: '活跃 Seeds',
        spider_seed_by_name: '按名称抓取',
        spider_refresh_oldest_detail: '刷新最久详情',
        spider_backfill_person: 'person 回填',
        spider_backfill_rank_period: '周期排行回填',
        spider_backfill_fetch_site: '外站任务回填',
        spider_fetch_javbus_resources: 'JavBus 资源抓取',
        spider_fetch_sukebei_resources: 'Sukebei 资源抓取',
        spider_rebuild_cast_rank: '演员 Rank 回填',
        spider_rebuild_actor_rank: '单演员 Rank',
        film_rename: '影片重命名',
        film_process: '影片处理',
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
        spider_rebuild_actor_rank: 'actor_name',
    };
    const STAGE_OVERVIEW_CONFIGS = {
        dailybest_stages: {
            phaseKeys: ['bestinv', 'detail', 'cover', 'translate'],
            labels: {
                bestinv: '榜单抓取',
                detail: '详情抓取',
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
                detailLoop: null,
                eventEntries: [],
                eventCount: 0,
                lastEventId: 0,
                eventSource: null,
                eventStreamKey: '',
                reconnectTimer: null,
                loadTimer: null,
                elapsedTimer: null,
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
        runtime.dailyBestBestinvProgressText = document.getElementById('dailyBestBestinvProgressText');
        runtime.dailyBestBestinvResultText = document.getElementById('dailyBestBestinvResultText');
        runtime.dailyBestDetailProgressText = document.getElementById('dailyBestDetailProgressText');
        runtime.dailyBestDetailResultText = document.getElementById('dailyBestDetailResultText');
        runtime.dailyBestCoverProgressText = document.getElementById('dailyBestCoverProgressText');
        runtime.dailyBestCoverResultText = document.getElementById('dailyBestCoverResultText');
        runtime.dailyBestTranslateProgressText = document.getElementById('dailyBestTranslateProgressText');
        runtime.dailyBestTranslateResultText = document.getElementById('dailyBestTranslateResultText');
        runtime.dailyBestStageCards = page.querySelectorAll('.overview-stage-card[data-phase]');
        runtime.nameFieldWrap = document.getElementById('nameFieldWrap');
        runtime.numberFieldWrap = document.getElementById('numberFieldWrap');
        runtime.actorNameFieldWrap = document.getElementById('actorNameFieldWrap');
        runtime.nameInput = document.getElementById('task_name');
        runtime.numberInput = document.getElementById('task_number');
        runtime.actorNameInput = document.getElementById('task_actor_name');

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

        runtime.syncTaskFields = function () {
            const taskType = String(runtime.taskTypeInput && runtime.taskTypeInput.value || '');
            runtime.toggleField(runtime.nameFieldWrap, taskType === 'spider_seed_by_name');
            runtime.toggleField(
                runtime.numberFieldWrap,
                taskType === 'spider_refresh_oldest_detail' ||
                taskType === 'spider_fetch_javbus_resources' ||
                taskType === 'spider_fetch_sukebei_resources'
            );
            runtime.toggleField(runtime.actorNameFieldWrap, taskType === 'spider_rebuild_actor_rank');
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
                taskType === 'spider_fetch_sukebei_resources'
            ) {
                payload.number = String(runtime.numberInput && runtime.numberInput.value || '').trim();
            }
            if (taskType === 'spider_rebuild_actor_rank') {
                payload.actor_name = String(runtime.actorNameInput && runtime.actorNameInput.value || '').trim();
            }

            if (requiredField && !String(payload[requiredField] || '').trim()) {
                throw new Error('请填写任务参数');
            }
            if (
                taskType === 'spider_refresh_oldest_detail' ||
                taskType === 'spider_fetch_javbus_resources' ||
                taskType === 'spider_fetch_sukebei_resources'
            ) {
                const parsed = Number(payload.number);
                if (!Number.isFinite(parsed)) {
                    throw new Error('请填写合法的数量');
                }
                payload.number = parsed;
            }
            return payload;
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
                if (source.readyState === EventSource.CLOSED) {
                    runtime.state.eventSource = null;
                    runtime.state.eventStreamKey = '';
                    runtime.state.reconnectTimer = global.setTimeout(function () {
                        runtime.state.reconnectTimer = null;
                        runtime.connectEventStream();
                    }, 1500);
                }
            };
        };

        runtime.updateControls = function () {
            const job = runtime.currentJob();
            const selectedId = runtime.normalizeJobId(runtime.state.selectedJobId);
            const paused = !!(job && job.paused);
            const done = !!(job && job.done);
            if (runtime.pauseBtn) runtime.pauseBtn.disabled = !job || paused || done;
            if (runtime.resumeBtn) runtime.resumeBtn.disabled = !job || !paused || done;
            if (runtime.stopBtn) runtime.stopBtn.disabled = (!job && !selectedId) || done;
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
                text: rawLine || (prefix + ' ' + String(event.message || '')),
            };
        };

        runtime.formatEventLineHTML = function (entry) {
            const raw = String((entry && entry.text) || '');
            const parts = raw.split(' | ');
            if (parts.length < 5 || !/^\[[^\]]+\]$/.test(parts[0] || '')) {
                return runtime.escapeHtml(raw).replace(/\b([A-Za-z0-9]+-\d{2,})\b/g, function (match, code) {
                    const href = '/movie/' + encodeURIComponent(code);
                    return '<a href="' + href + '" class="event-link">' + runtime.escapeHtml(code) + '</a>';
                });
            }
            const subjectId = String(parts[2] || '').trim();
            if (!/^\d+$/.test(subjectId) || String(parts[3] || '').indexOf('new_items=') !== 0) {
                return runtime.escapeHtml(raw).replace(/\b([A-Za-z0-9]+-\d{2,})\b/g, function (match, code) {
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
                current_phase_key: String(runtime.pickValue(event.current_phase_key, current.current_phase_key) || ''),
                phase_stats: runtime.pickValue(event.phase_stats, current.phase_stats) || null,
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
                current_phase_key: String(runtime.pickValue(event.current_phase_key, current.current_phase_key) || ''),
                phase_stats: runtime.pickValue(event.phase_stats, current.phase_stats) || null,
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

            if (runtime.form) {
                runtime.form.addEventListener('submit', function (event) {
                    event.preventDefault();
                    runtime.startTask();
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
