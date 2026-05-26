(function () {
    const API = {
        plan: '/api/triggers/sc-media/plan',
        precheck: '/api/triggers/sc-media/precheck',
        commit: '/api/triggers/sc-media/commit',
        back: '/api/triggers/sc-media/return',
    };

    const page = document.getElementById('scMediaMovePage');
    if (!page) {
        return;
    }

    const scopeModeInputs = Array.from(document.querySelectorAll('input[name="scopeMode"]'));
    const scNameGroup = document.getElementById('scNameGroup');
    const allModeHint = document.getElementById('allModeHint');
    const scNameInput = document.getElementById('scNameInput');
    const btnPrecheck = document.getElementById('btnPrecheck');
    const btnCommit = document.getElementById('btnCommit');
    const btnReturn = document.getElementById('btnReturn');
    const sumTotal = document.getElementById('sumTotal');
    const sumPass = document.getElementById('sumPass');
    const sumSkip = document.getElementById('sumSkip');
    const sumFail = document.getElementById('sumFail');
    const statusText = document.getElementById('statusText');
    const generatedAtText = document.getElementById('generatedAtText');
    const msgArea = document.getElementById('msgArea');

    let loading = false;
    let msgTimer = null;

    function currentMode() {
        const checked = scopeModeInputs.find(function (input) {
            return input && input.checked;
        });
        return checked ? String(checked.value || '').trim() : 'single';
    }

    function currentScName() {
        return scNameInput ? String(scNameInput.value || '').trim() : '';
    }

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
        return '<a href="/movie/' + encodeURIComponent(name) + '">' + escapeHtml(name) + '</a>';
    }

    function renderScLinks(scNames) {
        const names = Array.isArray(scNames) ? scNames.filter(function (name) {
            return String(name || '').trim() !== '';
        }) : [];
        if (!names.length) {
            return '-';
        }
        return names.map(function (name) {
            const safeName = String(name || '').trim();
            return '<a href="/sc-events/' + encodeURIComponent(safeName) + '">' + escapeHtml(safeName) + '</a>';
        }).join('<br>');
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
        scopeModeInputs.forEach(function (input) {
            input.disabled = loading;
        });
        if (scNameInput) {
            scNameInput.disabled = loading;
        }
    }

    function syncModeUi() {
        const mode = currentMode();
        const isAllMode = mode === 'all';
        if (scNameGroup) {
            scNameGroup.classList.toggle('d-none', isAllMode);
        }
        if (allModeHint) {
            allModeHint.classList.toggle('d-none', !isAllMode);
        }
    }

    function request(url, options) {
        return fetch(url, options).then(async function (resp) {
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

    function requestJSON(url, method, body) {
        return request(url, {
            method: method,
            headers: {'Content-Type': 'application/json'},
            body: body ? JSON.stringify(body) : null,
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
        if (sumPass) sumPass.textContent = String(safe.movable || 0);
        if (sumSkip) sumSkip.textContent = String(safe.skipped || 0);
        if (sumFail) sumFail.textContent = String(safe.failed || 0);
        if (statusText) statusText.textContent = safe.message || '等待操作';
        if (generatedAtText) {
            generatedAtText.textContent = safe.generated_at ? ('预处理时间：' + formatUnix(safe.generated_at)) : '';
        }
        if (btnCommit) {
            const canCommit = !!safe.can_commit;
            btnCommit.setAttribute('data-can-commit', canCommit ? '1' : '0');
            btnCommit.disabled = loading || !canCommit;
        }
        if (btnReturn) {
            if (safe.has_plan) {
                btnReturn.classList.remove('d-none');
            } else {
                btnReturn.classList.add('d-none');
            }
        }

        renderRows('tblPrecheckPass', safe.precheck_pass, [
            {render: function (r) { return renderScLinks(r.sc_names); }},
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.target_path || '-'); }},
        ]);

        renderRows('tblPrecheckSkip', safe.precheck_skip, [
            {render: function (r) { return renderScLinks(r.sc_names); }},
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {render: function (r) { return '<span class="text-muted">' + escapeHtml(r.error || '-') + '</span>'; }},
        ]);

        renderRows('tblPrecheckFail', safe.precheck_fail, [
            {render: function (r) { return renderScLinks(r.sc_names); }},
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {render: function (r) { return '<span class="text-danger">' + escapeHtml(r.error || '-') + '</span>'; }},
        ]);

        renderRows('tblCommitSuccess', safe.commit_success, [
            {render: function (r) { return renderScLinks(r.sc_names); }},
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.target_path || '-'); }},
        ]);

        renderRows('tblCommitFail', safe.commit_fail, [
            {render: function (r) { return renderScLinks(r.sc_names); }},
            {render: function (r) { return renderMovieLink(r.movie_name); }},
            {cls: 'mono', render: function (r) { return escapeHtml(r.source_path || '-'); }},
            {render: function (r) { return '<span class="text-danger">' + escapeHtml(r.error || '-') + '</span>'; }},
        ]);
    }

    function loadPlan() {
        const mode = currentMode();
        const scName = currentScName();
        if (mode === 'single' && !scName) {
            renderState({
                mode: mode,
                message: '请先输入 SC 名称',
                has_plan: false,
                can_commit: false,
                precheck_pass: [],
                precheck_skip: [],
                precheck_fail: [],
                commit_success: [],
                commit_fail: [],
            });
            return;
        }

        setLoading(true);
        request(API.plan + '?mode=' + encodeURIComponent(mode) + '&sc_name=' + encodeURIComponent(scName), {method: 'GET'})
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
            const mode = currentMode();
            const scName = currentScName();
            if (mode === 'single' && !scName) {
                showMsg('请先输入 SC 名称', false);
                return;
            }
            setLoading(true);
            requestJSON(API.precheck, 'POST', {mode: mode, sc_name: scName})
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
            const mode = currentMode();
            const scName = currentScName();
            if (mode === 'single' && !scName) {
                showMsg('请先输入 SC 名称', false);
                return;
            }
            setLoading(true);
            requestJSON(API.commit, 'POST', {mode: mode, sc_name: scName})
                .then(function (data) {
                    renderState(data);
                    showMsg(data.message || '第二段执行完成', true);
                })
                .catch(function (err) {
                    showMsg(err.message || '第二段执行失败', false);
                    if (err.result) {
                        renderState(err.result);
                    }
                })
                .finally(function () {
                    setLoading(false);
                });
        });
    }

    if (btnReturn) {
        btnReturn.addEventListener('click', function () {
            const mode = currentMode();
            const scName = currentScName();
            if (mode === 'single' && !scName) {
                showMsg('请先输入 SC 名称', false);
                return;
            }
            setLoading(true);
            requestJSON(API.back, 'POST', {mode: mode, sc_name: scName})
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

    scopeModeInputs.forEach(function (input) {
        input.addEventListener('change', function () {
            syncModeUi();
            loadPlan();
        });
    });

    syncModeUi();
    if (currentMode() === 'all' || currentScName()) {
        loadPlan();
    }
})();
