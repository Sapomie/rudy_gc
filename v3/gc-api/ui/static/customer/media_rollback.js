(function () {
    const API = {
        rollback: '/api/triggers/media/rollback',
    };

    const btnRollback = document.getElementById('btnRollback');
    const sumTotal = document.getElementById('sumTotal');
    const sumSuccess = document.getElementById('sumSuccess');
    const sumFail = document.getElementById('sumFail');
    const statusText = document.getElementById('statusText');
    const msgArea = document.getElementById('msgArea');

    let loading = false;
    let msgTimer = null;

    function escapeHtml(input) {
        return String(input || '')
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
        const href = '/movie/' + encodeURIComponent(name);
        return '<a href="' + href + '">' + escapeHtml(name) + '</a>';
    }

    function showMsg(text, ok) {
        if (!msgArea) {
            return;
        }
        clearTimeout(msgTimer);
        msgArea.textContent = text;
        msgArea.className = 'alert ' + (ok ? 'alert-success' : 'alert-danger') + ' fade show small py-2 px-3';
        msgArea.style.display = 'block';
        msgTimer = setTimeout(function () {
            msgArea.style.display = 'none';
        }, 3000);
    }

    function setLoading(next) {
        loading = !!next;
        if (btnRollback) {
            btnRollback.disabled = loading;
        }
    }

    function requestRollback() {
        return fetch(API.rollback, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
        }).then(async function (resp) {
            const payload = await resp.json().catch(function () {
                return {};
            });
            if (!resp.ok) {
                const result = payload && payload.result ? payload.result : null;
                const err = new Error((payload && payload.error) || ('请求失败(' + resp.status + ')'));
                err.result = result;
                throw err;
            }
            return payload || {};
        });
    }

    function renderRows(tableId, rows, columns) {
        const table = document.getElementById(tableId);
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
            const cells = columns.map(function (column) {
                return '<td class="' + (column.cls || '') + '">' + column.render(row) + '</td>';
            }).join('');
            return '<tr>' + cells + '</tr>';
        }).join('');
        tbody.innerHTML = html;
    }

    function renderState(data) {
        const safe = data || {};
        if (sumTotal) sumTotal.textContent = String(safe.total || 0);
        if (sumSuccess) sumSuccess.textContent = String(safe.success || 0);
        if (sumFail) sumFail.textContent = String(safe.failed || 0);
        if (statusText) statusText.textContent = safe.message || '等待操作';

        renderRows('tblRollbackSuccess', safe.rollback_success, [
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.target_path || '-'); }},
        ]);

        renderRows('tblRollbackFail', safe.rollback_fail, [
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {render: function (r) { return '<span class="text-danger">' + escapeHtml(r.error || '-') + '</span>'; }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.failed_path || '-'); }},
        ]);

        renderRows('tblRollbackPreview', safe.rollback_preview, [
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.target_path || '-'); }},
        ]);
    }

    if (btnRollback) {
        btnRollback.addEventListener('click', function () {
            setLoading(true);
            requestRollback()
                .then(function (data) {
                    renderState(data);
                    showMsg(data.message || '回滚完成', true);
                })
                .catch(function (err) {
                    showMsg(err.message || '回滚失败', false);
                    if (err.result) {
                        renderState(err.result);
                    }
                })
                .finally(function () {
                    setLoading(false);
                });
        });
    }
})();
