(function () {
    const page = document.getElementById('fetchSiteListPage');
    const msgEl = document.getElementById('fetchSiteListMsg');

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

    function startTask(taskType, movieJavID, movieCode, button) {
        setButtonState(button, true);
        request('/api/crawler/jobs/start', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                task_type: taskType,
                movie_jav_id: movieJavID,
                movie_code: movieCode
            })
        }).then(function () {
            const siteLabel = taskType === 'spider_fetch_javbus_resources' ? 'JavBus' : 'Sukebei';
            showMsg('已开启' + siteLabel + '抓取：' + movieCode, true);
            setButtonState(button, false);
        }).catch(function (error) {
            showMsg(error && error.message ? error.message : '启动失败', false);
            setButtonState(button, false);
        });
    }

    function readMovieTarget(button) {
        if (!button) {
            return {movieJavID: '', movieCode: ''};
        }

        const row = button.closest('tr[data-movie-jav-id]');
        if (row) {
            return {
                movieJavID: String(row.getAttribute('data-movie-jav-id') || '').trim(),
                movieCode: String(row.getAttribute('data-movie-code') || '').trim()
            };
        }

        const container = button.closest('[data-movie-jav-id][data-movie-code]') || page;
        if (!container) {
            return {movieJavID: '', movieCode: ''};
        }

        return {
            movieJavID: String(container.getAttribute('data-movie-jav-id') || '').trim(),
            movieCode: String(container.getAttribute('data-movie-code') || '').trim()
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
        if (!javbusButton && !sukebeiButton) {
            return;
        }

        const button = javbusButton || sukebeiButton;
        const target = readMovieTarget(button);
        const movieJavID = target.movieJavID;
        const movieCode = target.movieCode;
        if (!movieJavID || !movieCode) {
            showMsg('缺少影片参数，无法触发任务', false);
            return;
        }

        if (javbusButton) {
            startTask('spider_fetch_javbus_resources', movieJavID, movieCode, javbusButton);
            return;
        }
        startTask('spider_fetch_sukebei_resources', movieJavID, movieCode, sukebeiButton);
    });
})();
