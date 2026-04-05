(function () {
    const page = document.getElementById('fetchSiteListPage');
    const msgEl = document.getElementById('fetchSiteListMsg');
    const taskPageUrl = String(page && page.getAttribute('data-task-page-url') || '/triggers/fetch-site').trim() || '/triggers/fetch-site';
    const filteredTaskType = String(page && (page.getAttribute('data-filtered-task-type') || page.getAttribute('data-filtered-sukebei-task-type')) || '').trim();
    const filterFormID = String(page && page.getAttribute('data-filter-form-id') || '').trim();
    const fetchSiteFilterForm = document.getElementById(filterFormID || 'fetchSiteSukebeiFilterForm') || document.querySelector('form[data-fetch-site-filter-form]');

    if (!page) {
        return;
    }

    function showMsg(text, ok) {
        if (!msgEl) {
            return;
        }
        msgEl.textContent = String(text || '');
        msgEl.className = 'alert small py-2 px-3 mb-3 ' + (ok ? 'alert-success' : 'alert-danger');
    }

    function setButtonState(button, loading) {
        if (!button) {
            return;
        }
        button.disabled = loading;
    }

    function setTriggerButtonsLoading(loading) {
        document.querySelectorAll('.js-trigger-javbus, .js-trigger-sukebei, .js-trigger-both, .js-trigger-filtered-fetch-site, .js-trigger-filtered-sukebei').forEach(function (button) {
            setButtonState(button, loading);
        });
    }

    function request(url, options) {
        return fetch(url, options).then(async function (response) {
            const data = await response.json().catch(function () {
                return {};
            });
            if (!response.ok) {
                throw new Error(data.error || ('请求失败(' + response.status + ')'));
            }
            return data;
        });
    }

    function requestStartTaskPayload(payload) {
        return request('/api/crawler/jobs/start', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(payload)
        });
    }

    function requestStartTask(taskType, movieJavID, movieName) {
        return requestStartTaskPayload({
            task_type: taskType,
            movie_jav_id: movieJavID,
            movie_name: movieName
        });
    }

    function siteLabelByTaskType(taskType) {
        return taskType === 'spider_fetch_javbus_resources' ? 'JavBus' : 'Sukebei';
    }

    function startTask(taskType, movieJavID, movieName) {
        setTriggerButtonsLoading(true);
        requestStartTask(taskType, movieJavID, movieName).then(function () {
            showMsg('已开启' + siteLabelByTaskType(taskType) + '抓取：' + movieName, true);
        }).catch(function (error) {
            showMsg(error && error.message ? error.message : '启动失败', false);
        }).finally(function () {
            setTriggerButtonsLoading(false);
        });
    }

    function startBothTasks(movieJavID, movieName) {
        setTriggerButtonsLoading(true);
        Promise.allSettled([
            requestStartTask('spider_fetch_javbus_resources', movieJavID, movieName),
            requestStartTask('spider_fetch_sukebei_resources', movieJavID, movieName)
        ]).then(function (results) {
            const successSites = [];
            const failedSites = [];
            results.forEach(function (result, index) {
                const siteLabel = index === 0 ? 'JavBus' : 'Sukebei';
                if (result.status === 'fulfilled') {
                    successSites.push(siteLabel);
                    return;
                }
                const reason = result.reason && result.reason.message ? result.reason.message : '启动失败';
                failedSites.push(siteLabel + '：' + reason);
            });

            if (successSites.length === 2) {
                showMsg('已同时开启 JavBus + Sukebei 抓取：' + movieName, true);
                return;
            }
            if (successSites.length === 1) {
                showMsg('部分成功：已开启' + successSites[0] + '；' + failedSites.join('；'), false);
                return;
            }
            showMsg('同时触发失败：' + failedSites.join('；'), false);
        }).finally(function () {
            setTriggerButtonsLoading(false);
        });
    }

    function buildFilteredFetchSiteTaskPayload() {
        if (!filteredTaskType) {
            throw new Error('缺少筛选任务类型，无法触发');
        }
        if (!fetchSiteFilterForm) {
            throw new Error('缺少筛选表单，无法触发');
        }

        const payload = {task_type: filteredTaskType};
        const formData = new FormData(fetchSiteFilterForm);
        [
            'keyword',
            'last_fetch_from',
            'last_fetch_to',
            'release_date_from',
            'release_date_to',
            'film_birth_from',
            'film_birth_to',
            'media_birth_from',
            'media_birth_to',
            'owned',
            'mowned',
            'sort',
            'order'
        ].forEach(function (key) {
            const value = String(formData.get(key) || '').trim();
            if (value !== '') {
                payload[key] = value;
            }
        });
        const statuses = formData.getAll('statuses').map(function (value) {
            return String(value || '').trim();
        }).filter(function (value) {
            return value !== '';
        });
        if (statuses.length > 0) {
            payload.statuses = statuses;
        }
        const triggerSortKey = String(formData.get('trigger_sort_key') || '').trim();
        if (triggerSortKey !== '') {
            const parts = triggerSortKey.split(':');
            const triggerSort = String(parts[0] || '').trim();
            const triggerOrder = String(parts[1] || '').trim();
            if (triggerSort !== '') {
                payload.trigger_sort = triggerSort;
            }
            if (triggerOrder !== '') {
                payload.trigger_order = triggerOrder;
            }
        }
        return payload;
    }

    function startFilteredFetchSiteTask() {
        let payload;
        try {
            payload = buildFilteredFetchSiteTaskPayload();
        } catch (error) {
            showMsg(error && error.message ? error.message : '构造筛选任务失败', false);
            return;
        }

        setTriggerButtonsLoading(true);
        requestStartTaskPayload(payload).then(function (data) {
            const jobID = String(data && data.job_id || '').trim();
            if (!jobID) {
                throw new Error('任务已创建，但缺少 job_id');
            }
            window.location.href = taskPageUrl + '?job_id=' + encodeURIComponent(jobID);
        }).catch(function (error) {
            showMsg(error && error.message ? error.message : '启动失败', false);
            setTriggerButtonsLoading(false);
        });
    }

    function refreshSukebeiStatusClearButton() {
        const clearButton = document.querySelector('[data-status-clear]');
        if (!clearButton) {
            return;
        }
        const checkedCount = fetchSiteFilterForm ? fetchSiteFilterForm.querySelectorAll('input[name="statuses"]:checked').length : 0;
        clearButton.classList.toggle('active', checkedCount === 0);
    }

    function resolveAlbumName(button, fallback) {
        const raw = button ? String(button.getAttribute('data-album-name') || '').trim() : '';
        if (raw) {
            return raw;
        }
        const base = String(fallback || '').trim();
        return base || '下载中';
    }

    function applyFavoriteButtonState(button, favorited, albumName) {
        if (!button) {
            return;
        }
        const targetAlbumName = resolveAlbumName(button, albumName);
        button.setAttribute('data-favorited', favorited ? '1' : '0');
        button.setAttribute('data-album-name', targetAlbumName);
        button.classList.remove('btn-warning', 'btn-outline-warning', 'btn-info', 'btn-outline-info');

        const pendingAlbum = targetAlbumName === '待下载';
        if (pendingAlbum) {
            button.classList.add(favorited ? 'btn-info' : 'btn-outline-info');
        } else {
            button.classList.add(favorited ? 'btn-warning' : 'btn-outline-warning');
        }
        button.textContent = favorited ? ('已在' + targetAlbumName) : targetAlbumName;
        button.title = favorited ? ('移出' + targetAlbumName) : targetAlbumName;
    }

    function switchAlbumFavorite(movieJavID, sourceType, sourceRowID, currentFavorited, button, albumName) {
        const targetAlbumName = resolveAlbumName(button, albumName);
        const method = currentFavorited ? 'DELETE' : 'POST';
        setButtonState(button, true);
        request('/api/movie/' + encodeURIComponent(movieJavID) + '/album-item', {
            method: method,
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                movie_jav_id: movieJavID,
                source_type: sourceType,
                source_row_id: sourceRowID,
                album_name: targetAlbumName
            })
        }).then(function (data) {
            const favorited = data && typeof data.favorited === 'boolean' ? data.favorited : !currentFavorited;
            const responseAlbumName = data && data.album_name ? String(data.album_name).trim() : targetAlbumName;
            applyFavoriteButtonState(button, favorited, responseAlbumName);
            const message = data && data.message ? String(data.message) : '操作成功';
            showMsg(message, true);
            setButtonState(button, false);
        }).catch(function (error) {
            showMsg(error && error.message ? error.message : (targetAlbumName + '操作失败'), false);
            setButtonState(button, false);
        });
    }

    function readMovieTarget(button) {
        if (!button) {
            return {movieJavID: '', movieName: ''};
        }

        const row = button.closest('tr[data-movie-jav-id]');
        if (row) {
            return {
                movieJavID: String(row.getAttribute('data-movie-jav-id') || '').trim(),
                movieName: String(row.getAttribute('data-movie-name') || '').trim()
            };
        }

        const container = button.closest('[data-movie-jav-id][data-movie-name]') || page;
        if (!container) {
            return {movieJavID: '', movieName: ''};
        }

        return {
            movieJavID: String(container.getAttribute('data-movie-jav-id') || '').trim(),
            movieName: String(container.getAttribute('data-movie-name') || '').trim()
        };
    }

    function parseSizeToBytes(text) {
        const raw = String(text || '').trim().toUpperCase();
        const match = raw.match(/([0-9]+(?:\.[0-9]+)?)\s*(B|KB|KIB|MB|MIB|GB|GIB|TB|TIB)/);
        if (!match) {
            return 0;
        }
        const size = parseFloat(match[1]);
        const unit = match[2];
        const units = {
            B: 1,
            KB: 1024,
            KIB: 1024,
            MB: 1024 * 1024,
            MIB: 1024 * 1024,
            GB: 1024 * 1024 * 1024,
            GIB: 1024 * 1024 * 1024,
            TB: 1024 * 1024 * 1024 * 1024,
            TIB: 1024 * 1024 * 1024 * 1024
        };
        return size * (units[unit] || 1);
    }

    function parseDateValue(text) {
        const value = String(text || '').trim();
        const ts = Date.parse(value.replace(/\s+/g, 'T'));
        return Number.isNaN(ts) ? 0 : ts;
    }

    function parseNumberValue(text) {
        const value = parseInt(String(text || '').replace(/[^\d-]/g, ''), 10);
        return Number.isNaN(value) ? 0 : value;
    }

    function normalizeHashCells() {
        document.querySelectorAll('td.fetch-site-col-hash').forEach(function (cell) {
            if (cell.querySelector('.fetch-site-hash-text')) {
                return;
            }
            const value = String(cell.textContent || '').trim();
            cell.textContent = '';
            const span = document.createElement('span');
            span.className = 'fetch-site-hash-text';
            span.title = value;
            span.textContent = value;
            cell.appendChild(span);
        });
    }

    function extractSortValue(row, sortKey) {
        const cell = row.querySelector('[data-sort-key="' + sortKey + '"]');
        if (!cell) {
            return 0;
        }
        const text = cell.textContent || '';
        if (sortKey === 'size') {
            return parseSizeToBytes(text);
        }
        if (sortKey === 'date') {
            return parseDateValue(text);
        }
        if (sortKey === 'downloads') {
            return parseNumberValue(text);
        }
        return String(text || '').trim().toLowerCase();
    }

    function sortTable(tableID, sortKey, button) {
        const table = document.querySelector('[data-sort-table="' + tableID + '"]');
        if (!table) {
            return;
        }
        const tbody = table.querySelector('tbody');
        if (!tbody) {
            return;
        }
        const rows = Array.from(tbody.querySelectorAll('tr'));
        const currentKey = table.getAttribute('data-sort-current-key') || '';
        const currentOrder = table.getAttribute('data-sort-current-order') || 'desc';
        const nextOrder = currentKey === sortKey && currentOrder === 'desc' ? 'asc' : 'desc';

        rows.sort(function (left, right) {
            const leftValue = extractSortValue(left, sortKey);
            const rightValue = extractSortValue(right, sortKey);
            if (leftValue === rightValue) {
                return 0;
            }
            if (nextOrder === 'asc') {
                return leftValue > rightValue ? 1 : -1;
            }
            return leftValue < rightValue ? 1 : -1;
        });

        rows.forEach(function (row) {
            tbody.appendChild(row);
        });

        table.setAttribute('data-sort-current-key', sortKey);
        table.setAttribute('data-sort-current-order', nextOrder);

        document.querySelectorAll('.js-sort-table[data-table-id="' + tableID + '"]').forEach(function (item) {
            item.textContent = item.textContent.replace(/\s*[↑↓]$/, '');
        });
        if (button) {
            button.textContent = button.textContent.replace(/\s*[↑↓]$/, '') + (nextOrder === 'asc' ? ' ↑' : ' ↓');
        }
    }

    normalizeHashCells();

    document.addEventListener('click', function (event) {
        const sortButton = event.target.closest('.js-sort-table');
        if (sortButton) {
            sortTable(
                String(sortButton.getAttribute('data-table-id') || '').trim(),
                String(sortButton.getAttribute('data-sort-key') || '').trim(),
                sortButton
            );
            return;
        }

        const javbusButton = event.target.closest('.js-trigger-javbus');
        const sukebeiButton = event.target.closest('.js-trigger-sukebei');
        const bothButton = event.target.closest('.js-trigger-both');
        const filteredFetchSiteButton = event.target.closest('.js-trigger-filtered-fetch-site') || event.target.closest('.js-trigger-filtered-sukebei');
        const clearStatusButton = event.target.closest('[data-status-clear]');
        const favoriteButton = event.target.closest('.js-add-favorite');
        if (!javbusButton && !sukebeiButton && !bothButton && !filteredFetchSiteButton && !clearStatusButton && !favoriteButton) {
            return;
        }

        if (clearStatusButton) {
            if (fetchSiteFilterForm) {
                fetchSiteFilterForm.querySelectorAll('input[name="statuses"]').forEach(function (input) {
                    input.checked = false;
                });
            }
            refreshSukebeiStatusClearButton();
            return;
        }

        if (favoriteButton) {
            const target = readMovieTarget(favoriteButton);
            const movieJavID = target.movieJavID;
            const sourceType = String(favoriteButton.getAttribute('data-source-type') || '').trim();
            const sourceRowID = parseInt(String(favoriteButton.getAttribute('data-source-row-id') || '0'), 10);
            const currentFavorited = String(favoriteButton.getAttribute('data-favorited') || '').trim() === '1';
            const albumName = resolveAlbumName(favoriteButton);
            if (!movieJavID || !sourceType || !Number.isInteger(sourceRowID) || sourceRowID <= 0) {
                showMsg('缺少' + albumName + '参数，无法执行', false);
                return;
            }
            switchAlbumFavorite(movieJavID, sourceType, sourceRowID, currentFavorited, favoriteButton, albumName);
            return;
        }

        if (filteredFetchSiteButton) {
            startFilteredFetchSiteTask();
            return;
        }

        const button = javbusButton || sukebeiButton || bothButton;
        const target = readMovieTarget(button);
        const movieJavID = target.movieJavID;
        const movieName = target.movieName;
        if (!movieJavID || !movieName) {
            showMsg('缺少影片参数，无法触发任务', false);
            return;
        }

        if (bothButton) {
            startBothTasks(movieJavID, movieName);
            return;
        }

        if (javbusButton) {
            startTask('spider_fetch_javbus_resources', movieJavID, movieName);
            return;
        }
        startTask('spider_fetch_sukebei_resources', movieJavID, movieName);
    });

    if (fetchSiteFilterForm) {
        fetchSiteFilterForm.querySelectorAll('input[name="statuses"]').forEach(function (input) {
            input.addEventListener('change', refreshSukebeiStatusClearButton);
        });
    }

    refreshSukebeiStatusClearButton();
})();
