(function () {
    const API = {
        plan: '/api/triggers/media/plan',
        precheck: '/api/triggers/media/precheck',
        commit: '/api/triggers/media/commit',
        back: '/api/triggers/media/return',
    };

    const page = document.getElementById('mediaIngestPage');
    if (!page) {
        return;
    }

    const btnPrecheck = document.getElementById('btnPrecheck');
    const btnCommit = document.getElementById('btnCommit');
    const btnReturn = document.getElementById('btnReturn');
    const sumTotal = document.getElementById('sumTotal');
    const sumPassLabel = document.getElementById('sumPassLabel');
    const sumPass = document.getElementById('sumPass');
    const sumFailLabel = document.getElementById('sumFailLabel');
    const sumFail = document.getElementById('sumFail');
    const statusText = document.getElementById('statusText');
    const generatedAtText = document.getElementById('generatedAtText');
    const msgArea = document.getElementById('msgArea');

    let msgTimer = null;
    let loading = false;

    function escapeHtml(input) {
        return String(input || '')
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
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

    function renderMovieLink(movieName) {
        const name = String(movieName || '').trim();
        if (!name) {
            return '-';
        }
        const href = '/movie/' + encodeURIComponent(name);
        return '<a href="' + href + '">' + escapeHtml(name) + '</a>';
    }

    function formatUnix(ts) {
        const value = Number(ts || 0);
        if (!value) {
            return '';
        }
        const d = new Date(value * 1000);
        if (Number.isNaN(d.getTime())) {
            return '';
        }
        const y = d.getFullYear();
        const m = String(d.getMonth() + 1).padStart(2, '0');
        const day = String(d.getDate()).padStart(2, '0');
        const hh = String(d.getHours()).padStart(2, '0');
        const mm = String(d.getMinutes()).padStart(2, '0');
        const ss = String(d.getSeconds()).padStart(2, '0');
        return y + '-' + m + '-' + day + ' ' + hh + ':' + mm + ':' + ss;
    }

    function humanSize(bytes) {
        const value = Number(bytes || 0);
        if (!value) {
            return '-';
        }
        if (value < 1024) {
            return value + ' B';
        }
        if (value < 1024 * 1024) {
            return (value / 1024).toFixed(1) + ' KB';
        }
        if (value < 1024 * 1024 * 1024) {
            return (value / (1024 * 1024)).toFixed(1) + ' MB';
        }
        return (value / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
    }

    function isCommitPhase(data) {
        return String((data && data.phase) || '').trim() === 'commit';
    }

    function focusCommitResults(data) {
        const safe = data || {};
        const targetId = Array.isArray(safe.commit_fail) && safe.commit_fail.length
            ? 'commitFailSection'
            : (Array.isArray(safe.commit_success) && safe.commit_success.length ? 'commitSuccessSection' : 'commitPreviewSection');
        const target = document.getElementById(targetId);
        if (target && typeof target.scrollIntoView === 'function') {
            target.scrollIntoView({behavior: 'smooth', block: 'start'});
        }
    }

    function setLoading(next) {
        loading = !!next;
        if (btnPrecheck) {
            btnPrecheck.disabled = loading;
        }
        if (btnCommit) {
            const canCommit = btnCommit.getAttribute('data-can-commit') === '1';
            btnCommit.disabled = loading || !canCommit;
        }
        if (btnReturn) {
            btnReturn.disabled = loading;
        }
    }

    function request(url, method) {
        return fetch(url, {
            method: method || 'GET',
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
        const precheck = safe.precheck || {total: 0, passed: 0, failed: 0};
        const commitPhase = isCommitPhase(safe);
        const total = commitPhase ? Number(safe.total || 0) : Number(precheck.total || 0);
        const success = commitPhase ? Number(safe.success || 0) : Number(precheck.passed || 0);
        const failed = commitPhase ? Number(safe.failed || 0) : Number(precheck.failed || 0);

        if (sumPassLabel) sumPassLabel.textContent = commitPhase ? '成功' : '通过';
        if (sumFailLabel) sumFailLabel.textContent = '失败';
        if (sumTotal) sumTotal.textContent = String(total);
        if (sumPass) sumPass.textContent = String(success);
        if (sumFail) sumFail.textContent = String(failed);
        if (statusText) statusText.textContent = safe.message || '等待操作';
        if (generatedAtText) {
            if (commitPhase) {
                generatedAtText.textContent = '';
            } else {
                generatedAtText.textContent = safe.generated_at ? ('预处理时间：' + formatUnix(safe.generated_at)) : '';
            }
        }

        const canCommit = !!safe.can_commit;
        if (btnCommit) {
            btnCommit.setAttribute('data-can-commit', canCommit ? '1' : '0');
            btnCommit.disabled = loading || !canCommit;
            if (safe.partial_failed) {
                btnCommit.textContent = '第二段：仅插入通过项';
            } else {
                btnCommit.textContent = '第二段：执行插入';
            }
        }

        if (btnReturn) {
            if (safe.partial_failed && safe.has_plan) {
                btnReturn.classList.remove('d-none');
            } else {
                btnReturn.classList.add('d-none');
            }
        }

        renderRows('tblPrecheckPass', safe.precheck_pass, [
            {render: function (r) { return escapeHtml(r.movie_name || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.target_file_name || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.target_dir || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.alias || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_torrent_hash || '-'); }},
        ]);

        renderRows('tblPrecheckFail', safe.precheck_fail, [
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {render: function (r) { return '<span class="text-danger">' + escapeHtml(r.error || '-') + '</span>'; }},
        ]);

        renderRows('tblPrecheckPreview', safe.precheck_preview, [
            {render: function (r) { return escapeHtml(r.movie_name || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.target_dir || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.target_file_name || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.alias || '-'); }},
            {render: function (r) { return escapeHtml(humanSize(r.size)); }},
        ]);

        renderRows('tblCommitSuccess', safe.commit_success, [
            {render: function (r) { return escapeHtml(r.movie_name || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.target_path || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.alias || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_torrent_hash || '-'); }},
        ]);

        renderRows('tblCommitFail', safe.commit_fail, [
            {render: function (r) { return escapeHtml(r.movie_name || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {render: function (r) { return '<span class="text-danger">' + escapeHtml(r.error || '-') + '</span>'; }},
        ]);

        renderRows('tblCommitPreview', safe.commit_preview, [
            {render: function (r) { return escapeHtml(r.movie_name || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.target_path || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.alias || '-'); }},
            {render: function (r) { return escapeHtml(humanSize(r.size)); }},
        ]);
    }

    function loadPlan() {
        setLoading(true);
        request(API.plan, 'GET')
            .then(function (data) {
                renderState(data);
            })
            .catch(function (err) {
                showMsg(err.message || '读取预处理状态失败', false);
                if (err.result) {
                    renderState(err.result);
                }
            })
            .finally(function () {
                setLoading(false);
            });
    }

    if (btnPrecheck) {
        btnPrecheck.addEventListener('click', function () {
            setLoading(true);
            request(API.precheck, 'POST')
                .then(function (data) {
                    renderState(data);
                    showMsg(data.message || '预处理完成', true);
                })
                .catch(function (err) {
                    showMsg(err.message || '预处理失败', false);
                    if (err.result) {
                        renderState(err.result);
                    }
                })
                .finally(function () {
                    setLoading(false);
                });
        });
    }

    if (btnCommit) {
        btnCommit.addEventListener('click', function () {
            if (btnCommit.disabled) {
                return;
            }
            setLoading(true);
            if (statusText) {
                statusText.textContent = '正在执行第二段插入...';
            }
            request(API.commit, 'POST')
                .then(function (data) {
                    renderState(data);
                    focusCommitResults(data);
                    showMsg(data.message || '第二段执行完成', true);
                })
                .catch(function (err) {
                    showMsg(err.message || '第二段执行失败', false);
                    if (err.result) {
                        renderState(err.result);
                        focusCommitResults(err.result);
                    }
                })
                .finally(function () {
                    setLoading(false);
                });
        });
    }

    if (btnReturn) {
        btnReturn.addEventListener('click', function () {
            setLoading(true);
            request(API.back, 'POST')
                .then(function (data) {
                    renderState(data);
                    showMsg(data.message || '已返回', true);
                })
                .catch(function (err) {
                    showMsg(err.message || '返回失败', false);
                })
                .finally(function () {
                    setLoading(false);
                });
        });
    }

    loadPlan();
})();
