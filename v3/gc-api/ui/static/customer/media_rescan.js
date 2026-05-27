(function () {
    const API = {
        rescan: '/api/triggers/media/rescan',
    };

    const btnRescan = document.getElementById('btnRescan');
    const btnSelectAll = document.getElementById('btnSelectAll');
    const btnClearAll = document.getElementById('btnClearAll');
    const statusText = document.getElementById('statusText');
    const msgArea = document.getElementById('msgArea');

    const rootCheckboxes = Array.prototype.slice.call(document.querySelectorAll('.root-checkbox'));
    const branchCheckboxes = Array.prototype.slice.call(document.querySelectorAll('.branch-checkbox'));
    const rootScopeItems = Array.prototype.slice.call(document.querySelectorAll('.root-scope-item'));

    const sumTotalFiles = document.getElementById('sumTotalFiles');
    const sumMatched = document.getElementById('sumMatched');
    const sumMoved = document.getElementById('sumMoved');
    const sumRestored = document.getElementById('sumRestored');
    const sumRemoved = document.getElementById('sumRemoved');
    const sumUnmatched = document.getElementById('sumUnmatched');
    const sumSkippedRoots = document.getElementById('sumSkippedRoots');
    const sumErrors = document.getElementById('sumErrors');

    const selectedRootList = document.getElementById('selectedRootList');
    const scannedRootList = document.getElementById('scannedRootList');
    const skippedRootList = document.getElementById('skippedRootList');
    const selectedTargetList = document.getElementById('selectedTargetList');
    const scannedTargetList = document.getElementById('scannedTargetList');
    const skippedTargetList = document.getElementById('skippedTargetList');

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
        if (btnRescan) btnRescan.disabled = loading;
        if (btnSelectAll) btnSelectAll.disabled = loading;
        if (btnClearAll) btnClearAll.disabled = loading;
        rootCheckboxes.forEach(function (checkbox) {
            checkbox.disabled = loading;
        });
        branchCheckboxes.forEach(function (checkbox) {
            checkbox.disabled = loading || !isScopeChecked(checkbox.dataset.root, checkbox.dataset.scope);
        });
    }

    function isScopeChecked(root, scope) {
        return rootCheckboxes.some(function (checkbox) {
            return checkbox.dataset.root === root && checkbox.dataset.scope === scope && checkbox.checked;
        });
    }

    function syncBranchStates() {
        rootScopeItems.forEach(function (item) {
            const children = item.querySelectorAll('.branch-checkbox');
            children.forEach(function (checkbox) {
                const enabled = isScopeChecked(checkbox.dataset.root, checkbox.dataset.scope);
                checkbox.disabled = loading || !enabled;
                const label = checkbox.closest('.branch-checkbox-item');
                if (label) {
                    label.classList.toggle('is-disabled', !enabled);
                }
            });
        });
    }

    function buildSelections() {
        return rootCheckboxes.filter(function (checkbox) {
            return checkbox.checked;
        }).map(function (checkbox) {
            const root = String(checkbox.dataset.root || checkbox.value || '').trim();
            const scope = String(checkbox.dataset.scope || '').trim();
            const branches = branchCheckboxes.filter(function (branchCheckbox) {
                return branchCheckbox.dataset.root === root && branchCheckbox.dataset.scope === scope && branchCheckbox.checked;
            }).map(function (branchCheckbox) {
                return String(branchCheckbox.dataset.branch || '').trim();
            }).filter(function (value) {
                return value !== '';
            });
            return {
                root: root,
                scope: scope,
                branches: branches,
            };
        }).filter(function (item) {
            return item.root !== '' && item.scope !== '';
        });
    }

    function requestRescan(selections) {
        return fetch(API.rescan, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({selections: selections}),
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

    function renderList(target, items) {
        if (!target) {
            return;
        }
        const safeItems = Array.isArray(items) ? items : [];
        if (!safeItems.length) {
            target.innerHTML = '<li class="text-muted">暂无数据</li>';
            return;
        }
        target.innerHTML = safeItems.map(function (item) {
            return '<li><code>' + escapeHtml(item) + '</code></li>';
        }).join('');
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

        tbody.innerHTML = safeRows.map(function (row) {
            return '<tr>' + columns.map(function (column) {
                return '<td class="' + (column.cls || '') + '">' + column.render(row) + '</td>';
            }).join('') + '</tr>';
        }).join('');
    }

    function renderState(data) {
        const safe = data || {};
        if (sumTotalFiles) sumTotalFiles.textContent = String(safe.total_files || 0);
        if (sumMatched) sumMatched.textContent = String(safe.matched || 0);
        if (sumMoved) sumMoved.textContent = String(safe.moved || 0);
        if (sumRestored) sumRestored.textContent = String(safe.restored || 0);
        if (sumRemoved) sumRemoved.textContent = String(safe.marked_removed || 0);
        if (sumUnmatched) sumUnmatched.textContent = String(safe.unmatched || 0);
        if (sumSkippedRoots) sumSkippedRoots.textContent = String((safe.skipped_targets || []).length);
        if (sumErrors) sumErrors.textContent = String(safe.errors || 0);
        if (statusText) {
            statusText.textContent = buildStatusText(safe);
        }

        renderList(selectedRootList, safe.selected_roots);
        renderList(scannedRootList, safe.scanned_roots);
        renderList(skippedRootList, safe.skipped_roots);
        renderList(selectedTargetList, safe.selected_targets);
        renderList(scannedTargetList, safe.scanned_targets);
        renderList(skippedTargetList, safe.skipped_targets);

        renderRows('tblMoved', safe.moved_items, [
            {render: function (row) { return escapeHtml(row.movie_name || row.movie_jav_id || row.file_name || '-'); }},
            {render: function (row) { return escapeHtml(row.match_by || '-'); }},
            {cls: 'mono', render: function (row) { return escapeHtml(row.previous_path || '-'); }},
            {cls: 'mono', render: function (row) { return escapeHtml(row.path || '-'); }},
        ]);

        renderRows('tblRestored', safe.restored_items, [
            {render: function (row) { return escapeHtml(row.movie_name || row.movie_jav_id || row.file_name || '-'); }},
            {render: function (row) { return escapeHtml(row.match_by || '-'); }},
            {cls: 'mono', render: function (row) { return escapeHtml(row.previous_path || '-'); }},
            {cls: 'mono', render: function (row) { return escapeHtml(row.path || '-'); }},
        ]);

        renderRows('tblRemoved', safe.removed_items, [
            {render: function (row) { return escapeHtml(row.movie_name || row.movie_jav_id || row.file_name || '-'); }},
            {cls: 'mono', render: function (row) { return escapeHtml(row.root_dir || '-'); }},
            {cls: 'mono', render: function (row) { return escapeHtml(row.path || '-'); }},
        ]);

        renderRows('tblUnmatched', safe.unmatched_items, [
            {render: function (row) { return escapeHtml(row.file_name || '-'); }},
            {cls: 'mono', render: function (row) { return escapeHtml(row.root_dir || '-'); }},
            {cls: 'mono', render: function (row) { return escapeHtml(row.path || '-'); }},
        ]);

        renderRows('tblErrors', safe.error_items, [
            {render: function (row) { return escapeHtml(row.movie_name || row.file_name || '-'); }},
            {cls: 'mono', render: function (row) { return escapeHtml(row.path || '-'); }},
            {render: function (row) { return '<span class="text-danger">' + escapeHtml(row.error || '-') + '</span>'; }},
        ]);
    }

    function buildStatusText(data) {
        const safe = data || {};
        const scannedTargets = Array.isArray(safe.scanned_targets) ? safe.scanned_targets.length : 0;
        const skippedTargets = Array.isArray(safe.skipped_targets) ? safe.skipped_targets.length : 0;
        if (!safe.selected_roots || !safe.selected_roots.length) {
            return '等待操作';
        }
        if (scannedTargets === 0 && skippedTargets > 0) {
            return '所选范围当前都不存在，已全部跳过';
        }
        return '重扫完成：扫描 ' + scannedTargets + ' 个范围，跳过 ' + skippedTargets + ' 个范围';
    }

    if (btnSelectAll) {
        btnSelectAll.addEventListener('click', function () {
            rootCheckboxes.forEach(function (checkbox) {
                checkbox.checked = true;
            });
            branchCheckboxes.forEach(function (checkbox) {
                checkbox.checked = false;
            });
            syncBranchStates();
        });
    }

    if (btnClearAll) {
        btnClearAll.addEventListener('click', function () {
            rootCheckboxes.forEach(function (checkbox) {
                checkbox.checked = false;
            });
            branchCheckboxes.forEach(function (checkbox) {
                checkbox.checked = false;
            });
            syncBranchStates();
        });
    }

    rootCheckboxes.forEach(function (checkbox) {
        checkbox.addEventListener('change', syncBranchStates);
    });

    if (btnRescan) {
        btnRescan.addEventListener('click', function () {
            const selections = buildSelections();
            if (!selections.length) {
                showMsg('至少选择一个扫描目录', false);
                return;
            }

            setLoading(true);
            if (statusText) {
                statusText.textContent = '正在重扫...';
            }

            requestRescan(selections)
                .then(function (data) {
                    renderState(data);
                    showMsg('媒体重扫完成', true);
                })
                .catch(function (err) {
                    showMsg(err.message || '媒体重扫失败', false);
                    if (err.result) {
                        renderState(err.result);
                    }
                })
                .finally(function () {
                    setLoading(false);
                    syncBranchStates();
                });
        });
    }

    syncBranchStates();
})();
