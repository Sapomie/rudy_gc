(function () {
    const page = document.getElementById('albumItemsPage');
    const msgEl = document.getElementById('albumItemsMsg');
    const checkAll = document.getElementById('albumItemCheckAll');
    const filterForm = document.getElementById('albumItemsFilterForm');
    const albumNameSelect = document.getElementById('albumItemsAlbumName');
    const selectedCountEl = document.getElementById('albumBatchSelectedCount');
    const batchCopyBtn = document.getElementById('btnBatchCopyAlbumHashes');
    const batchMoveBtn = document.getElementById('btnBatchMoveAlbumItems');
    const batchRemoveBtn = document.getElementById('btnBatchRemoveAlbumItems');
    const batchMoveModalEl = document.getElementById('albumBatchMoveModal');
    const batchMoveTargetEl = document.getElementById('albumBatchMoveTarget');
    const batchMoveCountEl = document.getElementById('albumBatchMoveCount');
    const batchMoveConfirmBtn = document.getElementById('albumBatchMoveConfirmBtn');
    const batchRemoveModalEl = document.getElementById('albumBatchRemoveModal');
    const batchRemoveCountEl = document.getElementById('albumBatchRemoveCount');
    const batchRemoveConfirmBtn = document.getElementById('albumBatchRemoveConfirmBtn');
    if (!page) {
        return;
    }

    const albumID = parseInt(String(page.getAttribute('data-album-id') || '0'), 10);
    const batchMoveModal = (batchMoveModalEl && window.bootstrap && window.bootstrap.Modal)
        ? window.bootstrap.Modal.getOrCreateInstance(batchMoveModalEl)
        : null;
    const batchRemoveModal = (batchRemoveModalEl && window.bootstrap && window.bootstrap.Modal)
        ? window.bootstrap.Modal.getOrCreateInstance(batchRemoveModalEl)
        : null;
    let pendingBatchMoveIDs = [];
    let pendingBatchRemoveIDs = [];

    function showMsg(text, ok) {
        if (!msgEl) {
            return;
        }
        msgEl.textContent = String(text || '');
        msgEl.className = 'alert small py-2 px-3 mb-3 ' + (ok ? 'alert-success' : 'alert-danger');
    }

    function copyText(value) {
        const text = String(value || '').trim();
        if (!text) {
            return Promise.reject(new Error('缺少可复制内容'));
        }
        if (navigator.clipboard && navigator.clipboard.writeText) {
            return navigator.clipboard.writeText(text);
        }
        return new Promise(function (resolve, reject) {
            try {
                const input = document.createElement('textarea');
                input.value = text;
                input.style.position = 'fixed';
                input.style.left = '-9999px';
                document.body.appendChild(input);
                input.focus();
                input.select();
                const ok = document.execCommand('copy');
                document.body.removeChild(input);
                if (!ok) {
                    reject(new Error('复制失败'));
                    return;
                }
                resolve();
            } catch (error) {
                reject(error);
            }
        });
    }

    function request(url, options) {
        return fetch(url, options).then(async function (response) {
            const data = await response.json().catch(function () {
                return {};
            });
            if (!response.ok || !data || data.ok === false) {
                throw new Error((data && data.error) || ('请求失败(' + response.status + ')'));
            }
            return data;
        });
    }

    function parseItemID(raw) {
        const value = parseInt(String(raw || '0'), 10);
        if (!Number.isInteger(value) || value <= 0) {
            return 0;
        }
        return value;
    }

    function getCheckedItemIDs() {
        return Array.from(document.querySelectorAll('.js-album-item-check:checked')).map(function (checkbox) {
            return parseItemID(checkbox.value);
        }).filter(function (id) {
            return id > 0;
        });
    }

    function getCheckedHashes() {
        return Array.from(document.querySelectorAll('.js-album-item-check:checked')).map(function (checkbox) {
            const row = checkbox.closest('tr');
            if (!row) {
                return '';
            }
            const copyBtn = row.querySelector('.js-copy-hash');
            if (copyBtn) {
                return String(copyBtn.getAttribute('data-hash') || '').trim();
            }
            const hashCode = row.querySelector('.album-hash');
            return hashCode ? String(hashCode.textContent || '').trim() : '';
        }).filter(function (hash) {
            return hash.length > 0;
        });
    }

    function syncSelectionState() {
        const allChecks = Array.from(document.querySelectorAll('.js-album-item-check'));
        const checkedIDs = getCheckedItemIDs();
        if (selectedCountEl) {
            selectedCountEl.textContent = String(checkedIDs.length);
        }
        if (batchCopyBtn) {
            batchCopyBtn.disabled = checkedIDs.length === 0;
        }
        if (batchMoveBtn) {
            batchMoveBtn.disabled = checkedIDs.length === 0 || albumID <= 0;
        }
        if (batchRemoveBtn) {
            batchRemoveBtn.disabled = checkedIDs.length === 0 || albumID <= 0;
        }
        if (checkAll) {
            const total = allChecks.length;
            const checked = checkedIDs.length;
            checkAll.checked = total > 0 && checked === total;
            checkAll.indeterminate = checked > 0 && checked < total;
        }
    }

    function removeRowsByIDs(itemIDs) {
        itemIDs.forEach(function (id) {
            const row = page.querySelector('tr[data-item-id="' + id + '"]');
            if (row) {
                row.remove();
            }
        });
        syncSelectionState();
    }

    function removeAlbumItem(itemID, button) {
        if (albumID <= 0) {
            showMsg('当前相册无效，无法移除', false);
            return;
        }
        button.disabled = true;
        request('/api/albums/' + albumID + '/items/remove', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({item_id: itemID})
        }).then(function (data) {
            if (data.removed) {
                removeRowsByIDs([itemID]);
            }
            showMsg(data.message || '移除完成', true);
            button.disabled = false;
        }).catch(function (error) {
            showMsg(error && error.message ? error.message : '移除失败', false);
            button.disabled = false;
        });
    }

    function batchRemoveAlbumItems(itemIDs) {
        if (albumID <= 0) {
            showMsg('当前相册无效，无法批量移除', false);
            return;
        }
        if (!Array.isArray(itemIDs) || itemIDs.length === 0) {
            showMsg('请先选择要移除的条目', false);
            return;
        }
        if (batchRemoveBtn) {
            batchRemoveBtn.disabled = true;
        }
        request('/api/albums/' + albumID + '/items/batch-remove', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({item_ids: itemIDs})
        }).then(function (data) {
            const failedIDs = Array.isArray(data.failed_ids) ? data.failed_ids.map(parseItemID).filter(function (id) { return id > 0; }) : [];
            const failedMap = {};
            failedIDs.forEach(function (id) { failedMap[id] = true; });
            const removedIDs = itemIDs.filter(function (id) { return !failedMap[id]; });
            if (removedIDs.length > 0) {
                removeRowsByIDs(removedIDs);
            }
            if (failedIDs.length > 0) {
                showMsg('已移除 ' + (data.removed_count || removedIDs.length) + ' 条，失败 ' + failedIDs.length + ' 条', false);
            } else {
                showMsg('已批量移除 ' + (data.removed_count || removedIDs.length) + ' 条', true);
            }
            if (batchRemoveBtn) {
                batchRemoveBtn.disabled = false;
            }
        }).catch(function (error) {
            showMsg(error && error.message ? error.message : '批量移除失败', false);
            if (batchRemoveBtn) {
                batchRemoveBtn.disabled = false;
            }
        });
    }

    function batchMoveAlbumItems(itemIDs, targetAlbumID) {
        if (albumID <= 0) {
            showMsg('当前相册无效，无法批量移动', false);
            return;
        }
        if (!Array.isArray(itemIDs) || itemIDs.length === 0) {
            showMsg('请先选择要移动的条目', false);
            return;
        }
        if (targetAlbumID <= 0) {
            showMsg('请选择目标相册', false);
            return;
        }
        if (targetAlbumID === albumID) {
            showMsg('目标相册不能与当前相册相同', false);
            return;
        }
        if (batchMoveBtn) {
            batchMoveBtn.disabled = true;
        }
        request('/api/albums/' + albumID + '/items/batch-move', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({item_ids: itemIDs, target_album_id: targetAlbumID})
        }).then(function (data) {
            const failedIDs = Array.isArray(data.failed_ids) ? data.failed_ids.map(parseItemID).filter(function (id) { return id > 0; }) : [];
            const failedMap = {};
            failedIDs.forEach(function (id) { failedMap[id] = true; });
            const movedIDs = itemIDs.filter(function (id) { return !failedMap[id]; });
            if (movedIDs.length > 0) {
                removeRowsByIDs(movedIDs);
            }
            if (failedIDs.length > 0) {
                showMsg('已移动 ' + (data.moved_count || movedIDs.length) + ' 条，失败 ' + failedIDs.length + ' 条', false);
            } else {
                showMsg('已批量移动 ' + (data.moved_count || movedIDs.length) + ' 条', true);
            }
            if (batchMoveBtn) {
                batchMoveBtn.disabled = false;
            }
        }).catch(function (error) {
            showMsg(error && error.message ? error.message : '批量移动失败', false);
            if (batchMoveBtn) {
                batchMoveBtn.disabled = false;
            }
        });
    }

    function openBatchMoveModal(itemIDs) {
        if (!batchMoveModal || !batchMoveConfirmBtn || !batchMoveTargetEl) {
            showMsg('确认弹框未初始化，无法批量移动', false);
            return;
        }
        pendingBatchMoveIDs = itemIDs.slice();
        if (batchMoveCountEl) {
            batchMoveCountEl.textContent = String(pendingBatchMoveIDs.length);
        }
        batchMoveTargetEl.value = '';
        batchMoveModal.show();
    }

    function openBatchRemoveModal(itemIDs) {
        if (!batchRemoveModal || !batchRemoveModalEl || !batchRemoveConfirmBtn) {
            showMsg('确认弹框未初始化，无法批量移除', false);
            return;
        }
        pendingBatchRemoveIDs = itemIDs.slice();
        if (batchRemoveCountEl) {
            batchRemoveCountEl.textContent = String(pendingBatchRemoveIDs.length);
        }
        batchRemoveModal.show();
    }

    if (checkAll) {
        checkAll.addEventListener('change', function () {
            const checked = !!checkAll.checked;
            document.querySelectorAll('.js-album-item-check').forEach(function (checkbox) {
                checkbox.checked = checked;
            });
            syncSelectionState();
        });
    }

    if (batchRemoveBtn) {
        batchRemoveBtn.addEventListener('click', function () {
            const ids = getCheckedItemIDs();
            if (ids.length === 0) {
                showMsg('请先选择要移除的条目', false);
                return;
            }
            openBatchRemoveModal(ids);
        });
    }

    if (batchMoveBtn) {
        batchMoveBtn.addEventListener('click', function () {
            const ids = getCheckedItemIDs();
            if (ids.length === 0) {
                showMsg('请先选择要移动的条目', false);
                return;
            }
            openBatchMoveModal(ids);
        });
    }

    if (batchCopyBtn) {
        batchCopyBtn.addEventListener('click', function () {
            const hashes = getCheckedHashes();
            if (hashes.length === 0) {
                showMsg('请先选择要复制的条目', false);
                return;
            }
            copyText(hashes.join('\n')).then(function () {
                showMsg('已复制 ' + hashes.length + ' 条 Hash', true);
            }).catch(function (error) {
                showMsg(error && error.message ? error.message : '复制失败', false);
            });
        });
    }

    if (batchMoveModalEl) {
        batchMoveModalEl.addEventListener('hidden.bs.modal', function () {
            pendingBatchMoveIDs = [];
            if (batchMoveCountEl) {
                batchMoveCountEl.textContent = '0';
            }
            if (batchMoveTargetEl) {
                batchMoveTargetEl.value = '';
            }
            if (batchMoveConfirmBtn) {
                batchMoveConfirmBtn.disabled = false;
            }
        });
    }

    if (batchMoveConfirmBtn) {
        batchMoveConfirmBtn.addEventListener('click', function () {
            if (!Array.isArray(pendingBatchMoveIDs) || pendingBatchMoveIDs.length === 0) {
                showMsg('缺少待移动条目', false);
                if (batchMoveModal) {
                    batchMoveModal.hide();
                }
                return;
            }
            const targetAlbumID = parseItemID(batchMoveTargetEl ? batchMoveTargetEl.value : '');
            if (targetAlbumID <= 0) {
                showMsg('请选择目标相册', false);
                return;
            }
            if (targetAlbumID === albumID) {
                showMsg('目标相册不能与当前相册相同', false);
                return;
            }
            batchMoveConfirmBtn.disabled = true;
            const ids = pendingBatchMoveIDs.slice();
            batchMoveAlbumItems(ids, targetAlbumID);
            if (batchMoveModal) {
                batchMoveModal.hide();
            }
        });
    }

    if (batchRemoveModalEl) {
        batchRemoveModalEl.addEventListener('hidden.bs.modal', function () {
            pendingBatchRemoveIDs = [];
            if (batchRemoveCountEl) {
                batchRemoveCountEl.textContent = '0';
            }
            if (batchRemoveConfirmBtn) {
                batchRemoveConfirmBtn.disabled = false;
            }
        });
    }

    if (batchRemoveConfirmBtn) {
        batchRemoveConfirmBtn.addEventListener('click', function () {
            if (!Array.isArray(pendingBatchRemoveIDs) || pendingBatchRemoveIDs.length === 0) {
                showMsg('缺少待移除条目', false);
                if (batchRemoveModal) {
                    batchRemoveModal.hide();
                }
                return;
            }
            batchRemoveConfirmBtn.disabled = true;
            const ids = pendingBatchRemoveIDs.slice();
            batchRemoveAlbumItems(ids);
            if (batchRemoveModal) {
                batchRemoveModal.hide();
            }
        });
    }

    document.addEventListener('change', function (event) {
        const target = event.target;
        if (target && target.classList && target.classList.contains('js-album-item-check')) {
            syncSelectionState();
        }
    });

    document.addEventListener('click', function (event) {
        const quickFilterButton = event.target.closest('.js-album-quick-filter');
        if (quickFilterButton) {
            const targetAlbumName = String(quickFilterButton.getAttribute('data-album-name') || '').trim();
            if (!targetAlbumName) {
                showMsg('缺少相册参数，无法快捷筛选', false);
                return;
            }
            if (!filterForm || !albumNameSelect) {
                showMsg('筛选表单未初始化，无法快捷筛选', false);
                return;
            }
            albumNameSelect.value = targetAlbumName;
            filterForm.submit();
            return;
        }

        const copyButton = event.target.closest('.js-copy-hash');
        if (copyButton) {
            const hash = String(copyButton.getAttribute('data-hash') || '').trim();
            copyText(hash).then(function () {
                showMsg('Info Hash 已复制', true);
            }).catch(function (error) {
                showMsg(error && error.message ? error.message : '复制失败', false);
            });
            return;
        }

        const removeButton = event.target.closest('.js-remove-album-item');
        if (!removeButton) {
            return;
        }
        const itemID = parseItemID(removeButton.getAttribute('data-item-id'));
        if (itemID <= 0) {
            showMsg('缺少相册条目参数，无法移除', false);
            return;
        }
        removeAlbumItem(itemID, removeButton);
    });

    syncSelectionState();
})();
