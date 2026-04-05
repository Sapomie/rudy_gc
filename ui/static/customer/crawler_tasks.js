(function (global) {
    const page = document.getElementById('crawlerTasksPage');
    if (!page) {
        return;
    }

    const API_BASE = '/api/crawler/jobs';
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
        spider_fetch_sehuatang_magnets: '色花堂磁力抓取',
        spider_rebuild_cast_rank: '演员 Rank 回填',
        spider_rebuild_actor_rank: '单演员 Rank',
        film_rename: '影片重命名',
        film_process: '影片处理',
        sc_rebuild_stats: 'SC 统计回填',
        sc_move: 'SC 影片移动',
        sc_add: 'SC 新增',
    };
    const TASK_TRIGGER_PAGES = {
        spider_daily_best: '/triggers/dailybest',
        spider_daily_best_sync: '/triggers/dailybest',
        spider_seeds: '/triggers/seeds',
        spider_seed_by_name: '/triggers/seeds',
        spider_refresh_oldest_detail: '/triggers/seeds',
        spider_download_cover: '/triggers/post-process',
        spider_translate_title: '/triggers/post-process',
        spider_post_process: '/triggers/post-process',
        spider_backfill_person: '/triggers/backfill',
        spider_backfill_rank_period: '/triggers/backfill',
        spider_backfill_fetch_site: '/triggers/backfill',
        spider_fetch_javbus_resources: '/triggers/fetch-site',
        spider_fetch_javbus_filtered_resources: '/triggers/fetch-site-javbus-filtered',
        spider_fetch_sukebei_resources: '/triggers/fetch-site',
        spider_fetch_sukebei_filtered_resources: '/triggers/fetch-site-sukebei-filtered',
        spider_fetch_sehuatang_magnets: '/triggers/fetch-sehuatang',
        film_rename: '/triggers/film',
        film_process: '/triggers/film',
        spider_rebuild_cast_rank: '/triggers/backfill',
        spider_rebuild_actor_rank: '/triggers/backfill',
        sc_rebuild_stats: '/triggers/backfill',
        sc_move: '/triggers/backfill',
        sc_add: '/triggers/sc',
    };

    const tableBody = document.getElementById('taskTableBody');
    const taskCountBadge = document.getElementById('taskCountBadge');
    const triggerPageUrl = (global.CRAWLER_TASKS_PAGE_DATA && global.CRAWLER_TASKS_PAGE_DATA.triggerPageUrl) || '/triggers/dailybest';

    function escapeHtml(input) {
        return String(input || '')
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    function timeText(ts) {
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
    }

    function stageLabel(stage) {
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
                return current || '-';
        }
    }

    function taskLabel(taskType) {
        return TASK_LABELS[String(taskType || '').trim()] || String(taskType || '-');
    }

    function triggerPageForTask(taskType) {
        const current = String(taskType || '').trim();
        return TASK_TRIGGER_PAGES[current] || triggerPageUrl;
    }

    function renderRows(list) {
        const safeList = Array.isArray(list) ? list.slice() : [];
        safeList.sort(function (a, b) {
            const aStarted = Number((a || {}).started_at || 0);
            const bStarted = Number((b || {}).started_at || 0);
            if (aStarted !== bStarted) {
                return bStarted - aStarted;
            }
            const aID = String((a || {}).job_id || '');
            const bID = String((b || {}).job_id || '');
            if (aID > bID) return -1;
            if (aID < bID) return 1;
            return 0;
        });

        if (taskCountBadge) {
            taskCountBadge.textContent = String(safeList.length) + ' 个';
        }
        if (!safeList.length) {
            tableBody.innerHTML = '<tr><td colspan="8" class="text-center text-muted py-4">暂无任务</td></tr>';
            return;
        }

        tableBody.innerHTML = safeList.map(function (item) {
            const jobID = String(item.job_id || '');
            const message = item && item.message ? item.message : stageLabel(item && item.stage);
            const jumpUrl = triggerPageForTask(item.task_type) + '?job_id=' + encodeURIComponent(jobID);
            return '' +
                '<tr>' +
                '<td class="task-id-cell" title="' + escapeHtml(jobID) + '">' + escapeHtml(jobID || '-') + '</td>' +
                '<td>' + escapeHtml(taskLabel(item.task_type)) + '</td>' +
                '<td title="' + escapeHtml(message) + '">' + escapeHtml(message) + '</td>' +
                '<td>' + escapeHtml(String(item.queued_count || 0)) + '</td>' +
                '<td>' + escapeHtml(String(item.handled_count || 0)) + '</td>' +
                '<td>' + escapeHtml(String(item.success_count || 0)) + ' / ' + escapeHtml(String(item.failed_count || 0)) + '</td>' +
                '<td>' + escapeHtml(timeText(item.started_at)) + '</td>' +
                '<td><a class="btn btn-sm btn-primary" href="' + jumpUrl + '">回到任务</a></td>' +
                '</tr>';
        }).join('');
    }

    function loadTasks() {
        fetch(API_BASE, {method: 'GET'})
            .then(async function (response) {
                const data = await response.json().catch(function () {
                    return {};
                });
                if (!response.ok) {
                    throw new Error(data.error || ('请求失败(' + response.status + ')'));
                }
                return data;
            })
            .then(function (payload) {
                renderRows(payload.jobs || []);
            })
            .catch(function () {
                tableBody.innerHTML = '<tr><td colspan="8" class="text-center text-danger py-4">加载失败</td></tr>';
            });
    }

    loadTasks();
    global.setInterval(loadTasks, 5000);
})(window);
