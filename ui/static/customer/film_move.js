(function () {
    const API = {
        preview: '/api/triggers/film-move/preview',
        commit: '/api/triggers/film-move/commit',
    };

    const page = document.getElementById('filmMovePage');
    if (!page) {
        return;
    }

    const form = document.getElementById('filmMoveFilterForm');
    const btnPreview = document.getElementById('btnPreview');
    const btnCommit = document.getElementById('btnCommit');
    const statusText = document.getElementById('statusText');
    const planText = document.getElementById('planText');
    const sumTotal = document.getElementById('sumTotal');
    const sumMovable = document.getElementById('sumMovable');
    const sumPreviewFailed = document.getElementById('sumPreviewFailed');
    const sumMoved = document.getElementById('sumMoved');
    const sumMoveFailed = document.getElementById('sumMoveFailed');
    const msgArea = document.getElementById('msgArea');

    let loading = false;
    let msgTimer = null;
    let currentPlanID = '';

    function showMsg(text, ok) {
        if (!msgArea) {
            return;
        }
        clearTimeout(msgTimer);
        msgArea.textContent = String(text || '');
        msgArea.className = 'alert ' + (ok ? 'alert-success' : 'alert-danger') + ' fade show small py-2 px-3';
        msgArea.style.display = 'block';
        msgTimer = setTimeout(function () {
            msgArea.style.display = 'none';
        }, 3000);
    }

    function setStatus(text) {
        if (statusText) {
            statusText.textContent = text || '等待操作';
        }
    }

    function setPlanText(text) {
        if (planText) {
            planText.textContent = text || '';
        }
    }

    function escapeHtml(value) {
        return String(value || '')
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    function renderMovieLink(movieName) {
        const name = String(movieName || '').trim();
        if (!name) {
            return '-';
        }
        return '<a href="/movie/' + encodeURIComponent(name) + '">' + escapeHtml(name) + '</a>';
    }

    function toDirPath(pathValue) {
        const raw = String(pathValue || '').trim();
        if (!raw) {
            return '-';
        }
        const normalized = raw.replace(/[\\/]+$/, '');
        const slashIdx = Math.max(normalized.lastIndexOf('/'), normalized.lastIndexOf('\\'));
        if (slashIdx < 0) {
            return normalized;
        }
        if (slashIdx === 0) {
            return normalized.slice(0, 1);
        }
        return normalized.slice(0, slashIdx);
    }

    function setLoading(next) {
        loading = !!next;
        if (btnPreview) {
            btnPreview.disabled = loading;
        }
        if (btnCommit) {
            btnCommit.disabled = loading || currentPlanID === '';
        }
    }

    function shouldSkipElement(el) {
        if (!el || !el.name || el.disabled) {
            return true;
        }
        const type = String(el.type || '').toLowerCase();
        return type === 'submit' || type === 'button' || type === 'reset' || type === 'fieldset';
    }

    function normalizeValue(el) {
        return String(el && el.value ? el.value : '').trim();
    }

    function collectFilterParams() {
        const params = new URLSearchParams();
        if (!form) {
            return params;
        }

        Array.prototype.forEach.call(form.elements, function (el) {
            if (shouldSkipElement(el)) {
                return;
            }

            const type = String(el.type || '').toLowerCase();
            const name = String(el.name || '').trim();
            if (!name) {
                return;
            }

            if ((type === 'checkbox' || type === 'radio') && !el.checked) {
                return;
            }

            const value = normalizeValue(el);
            if (type === 'hidden') {
                const explicit = el.dataset.explicit === '1';
                if (explicit && value !== '') {
                    params.set(name, value);
                }
                return;
            }

            if (value === '') {
                return;
            }
            params.set(name, value);
        });

        return params;
    }

    function request(url, options) {
        return fetch(url, options).then(async function (resp) {
            const payload = await resp.json().catch(function () {
                return {};
            });
            if (!resp.ok) {
                throw new Error(payload.error || ('请求失败(' + resp.status + ')'));
            }
            return payload || {};
        });
    }

    function renderRows(tableID, rows, columns) {
        const table = document.getElementById(tableID);
        if (!table) {
            return;
        }
        const tbody = table.querySelector('tbody');
        if (!tbody) {
            return;
        }

        const safeRows = Array.isArray(rows) ? rows : [];
        if (!safeRows.length) {
            tbody.innerHTML = '<tr><td colspan="' + columns.length + '" class="text-center text-muted">暂无数据</td></tr>';
            return;
        }

        const html = safeRows.map(function (row) {
            const cells = columns.map(function (col) {
                return '<td class="' + (col.cls || '') + '">' + col.render(row) + '</td>';
            }).join('');
            return '<tr>' + cells + '</tr>';
        }).join('');
        tbody.innerHTML = html;
    }

    function resetCommitTables() {
        renderRows('tblCommitSuccess', [], [
            {render: function () { return ''; }},
            {render: function () { return ''; }},
            {render: function () { return ''; }},
        ]);
        renderRows('tblCommitFail', [], [
            {render: function () { return ''; }},
            {render: function () { return ''; }},
            {render: function () { return ''; }},
            {render: function () { return ''; }},
        ]);
        if (sumMoved) sumMoved.textContent = '0';
        if (sumMoveFailed) sumMoveFailed.textContent = '0';
    }

    function renderPreview(data) {
        const safe = data || {};
        currentPlanID = String(safe.plan_id || '').trim();

        if (sumTotal) sumTotal.textContent = String(safe.total || 0);
        if (sumMovable) sumMovable.textContent = String(safe.movable || 0);
        if (sumPreviewFailed) sumPreviewFailed.textContent = String(safe.failed || 0);

        setStatus(safe.message || '第一步完成');
        setPlanText(currentPlanID ? ('当前计划ID：' + currentPlanID) : '当前无可执行计划');

        if (btnCommit) {
            const movable = Number(safe.movable || 0);
            btnCommit.disabled = loading || currentPlanID === '' || movable <= 0;
        }

        renderRows('tblPreview', safe.items, [
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'path-cell', render: function (r) { return escapeHtml(toDirPath(r.source_path)); }},
            {cls: 'path-cell', render: function (r) { return escapeHtml(toDirPath(r.target_path)); }},
            {
                render: function (r) {
                    if (r.can_move) {
                        return '<span class="text-success">可移动</span>';
                    }
                    return '<span class="text-danger">' + escapeHtml(r.error || '不可移动') + '</span>';
                }
            },
        ]);

        resetCommitTables();
    }

    function renderCommit(data) {
        const safe = data || {};
        currentPlanID = '';
        if (btnCommit) {
            btnCommit.disabled = true;
        }

        if (sumMoved) sumMoved.textContent = String(safe.success || 0);
        if (sumMoveFailed) sumMoveFailed.textContent = String(safe.failed || 0);

        setStatus(safe.message || '第二步完成');
        setPlanText('本次计划已执行完成');

        renderRows('tblCommitSuccess', safe.success_items, [
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'path-cell', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {cls: 'path-cell', render: function (r) { return escapeHtml(r.target_path || '-'); }},
        ]);

        renderRows('tblCommitFail', safe.failed_items, [
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'path-cell', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {cls: 'path-cell', render: function (r) { return escapeHtml(r.target_path || '-'); }},
            {render: function (r) { return '<span class="text-danger">' + escapeHtml(r.error || '-') + '</span>'; }},
        ]);
    }

    if (btnPreview) {
        btnPreview.addEventListener('click', function () {
            const params = collectFilterParams();
            const query = params.toString();
            const url = query ? (API.preview + '?' + query) : API.preview;

            setLoading(true);
            request(url, {method: 'GET'})
                .then(function (data) {
                    renderPreview(data);
                    showMsg(data.message || '第一步完成', true);
                })
                .catch(function (err) {
                    showMsg(err && err.message ? err.message : '第一步执行失败', false);
                })
                .finally(function () {
                    setLoading(false);
                });
        });
    }

    if (btnCommit) {
        btnCommit.addEventListener('click', function () {
            if (!currentPlanID) {
                showMsg('请先执行第一步', false);
                return;
            }

            setLoading(true);
            request(API.commit, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({plan_id: currentPlanID}),
            })
                .then(function (data) {
                    renderCommit(data);
                    showMsg(data.message || '第二步完成', true);
                })
                .catch(function (err) {
                    showMsg(err && err.message ? err.message : '第二步执行失败', false);
                })
                .finally(function () {
                    setLoading(false);
                });
        });
    }
})();
