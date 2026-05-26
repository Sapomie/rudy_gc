(function () {
    const page = document.getElementById('torrentAlbumsPage');
    const msgEl = document.getElementById('torrentAlbumsMsg');
    const checkAll = document.getElementById('torrentAlbumItemCheckAll');
    const filterForm = document.getElementById('torrentAlbumsFilterForm');
    const albumNameSelect = document.getElementById('torrentAlbumsAlbumName');
    const selectedCountEl = document.getElementById('torrentAlbumBatchSelectedCount');
    const batchCopyBtn = document.getElementById('btnBatchCopyTorrentAlbumHashes');
    const batchMoveBtn = document.getElementById('btnBatchMoveTorrentAlbumItems');
    const batchRemoveBtn = document.getElementById('btnBatchRemoveTorrentAlbumItems');
    const batchMoveModalEl = document.getElementById('torrentAlbumBatchMoveModal');
    const batchMoveTargetEl = document.getElementById('torrentAlbumBatchMoveTarget');
    const batchMoveCountEl = document.getElementById('torrentAlbumBatchMoveCount');
    const batchMoveConfirmBtn = document.getElementById('torrentAlbumBatchMoveConfirmBtn');
    const batchRemoveModalEl = document.getElementById('torrentAlbumBatchRemoveModal');
    const batchRemoveCountEl = document.getElementById('torrentAlbumBatchRemoveCount');
    const batchRemoveConfirmBtn = document.getElementById('torrentAlbumBatchRemoveConfirmBtn');
    const openCreateTorrentAlbumBtn = document.getElementById('btnOpenCreateTorrentAlbumModal');
    const createTorrentAlbumModalEl = document.getElementById('torrentAlbumCreateModal');
    const createTorrentAlbumNameEl = document.getElementById('torrentAlbumCreateName');
    const createTorrentAlbumConfirmBtn = document.getElementById('torrentAlbumCreateConfirmBtn');
    if (!page) {
        return;
    }

    const torrentAlbumID = parseInt(String(page.getAttribute('data-album-id') || '0'), 10);
    const batchMoveModal = (batchMoveModalEl && window.bootstrap && window.bootstrap.Modal)
        ? window.bootstrap.Modal.getOrCreateInstance(batchMoveModalEl)
        : null;
    const batchRemoveModal = (batchRemoveModalEl && window.bootstrap && window.bootstrap.Modal)
        ? window.bootstrap.Modal.getOrCreateInstance(batchRemoveModalEl)
        : null;
    const createTorrentAlbumModal = (createTorrentAlbumModalEl && window.bootstrap && window.bootstrap.Modal)
        ? window.bootstrap.Modal.getOrCreateInstance(createTorrentAlbumModalEl)
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
        return Array.from(document.querySelectorAll('.js-torrent-album-item-check:checked')).map(function (checkbox) {
            return parseItemID(checkbox.value);
        }).filter(function (id) {
            return id > 0;
        });
    }

    function getCheckedHashes() {
        return Array.from(document.querySelectorAll('.js-torrent-album-item-check:checked')).map(function (checkbox) {
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
        const allChecks = Array.from(document.querySelectorAll('.js-torrent-album-item-check'));
        const checkedIDs = getCheckedItemIDs();
        if (selectedCountEl) {
            selectedCountEl.textContent = String(checkedIDs.length);
        }
        if (batchCopyBtn) {
            batchCopyBtn.disabled = checkedIDs.length === 0;
        }
        if (batchMoveBtn) {
            batchMoveBtn.disabled = checkedIDs.length === 0 || torrentAlbumID <= 0;
        }
        if (batchRemoveBtn) {
            batchRemoveBtn.disabled = checkedIDs.length === 0 || torrentAlbumID <= 0;
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

    function removeTorrentAlbumItem(itemID, button) {
        if (torrentAlbumID <= 0) {
            showMsg('当前相册无效，无法移除', false);
            return;
        }
        button.disabled = true;
        request('/api/albums/' + torrentAlbumID + '/items/remove', {
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

    function batchRemoveTorrentAlbumItems(itemIDs) {
        if (torrentAlbumID <= 0) {
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
        request('/api/albums/' + torrentAlbumID + '/items/batch-remove', {
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

    function batchMoveTorrentAlbumItems(itemIDs, targetAlbumID) {
        if (torrentAlbumID <= 0) {
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
        if (targetAlbumID === torrentAlbumID) {
            showMsg('目标相册不能与当前相册相同', false);
            return;
        }
        if (batchMoveBtn) {
            batchMoveBtn.disabled = true;
        }
        request('/api/albums/' + torrentAlbumID + '/items/batch-move', {
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

    function upsertSelectOption(selectEl, value, text) {
        if (!selectEl) {
            return;
        }
        const stringValue = String(value || '').trim();
        const stringText = String(text || '').trim();
        if (!stringValue || !stringText) {
            return;
        }
        let targetOption = null;
        Array.from(selectEl.options || []).forEach(function (option) {
            if (String(option.value || '').trim() === stringValue) {
                targetOption = option;
            }
        });
        if (!targetOption) {
            targetOption = document.createElement('option');
            targetOption.value = stringValue;
            targetOption.textContent = stringText;
            selectEl.appendChild(targetOption);
            return;
        }
        targetOption.textContent = stringText;
    }

    function submitFilterWithTorrentAlbumName(albumName) {
        const targetAlbumName = String(albumName || '').trim();
        if (!targetAlbumName || !filterForm || !albumNameSelect) {
            return;
        }
        upsertSelectOption(albumNameSelect, targetAlbumName, targetAlbumName);
        albumNameSelect.value = targetAlbumName;
        filterForm.submit();
    }

    function createTorrentAlbum(albumName) {
        const name = String(albumName || '').trim();
        if (!name) {
            showMsg('相册名称不能为空', false);
            return;
        }
        if (!createTorrentAlbumConfirmBtn) {
            showMsg('创建按钮未初始化，无法新增相册', false);
            return;
        }
        createTorrentAlbumConfirmBtn.disabled = true;
        request('/api/albums', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({name: name})
        }).then(function (data) {
            const newAlbumName = String(data && data.album_name ? data.album_name : '').trim();
            const newAlbumID = parseItemID(data && data.album_id ? data.album_id : '0');
            if (!newAlbumName || newAlbumID <= 0) {
                throw new Error('新增相册返回数据无效');
            }
            upsertSelectOption(albumNameSelect, newAlbumName, newAlbumName);
            upsertSelectOption(batchMoveTargetEl, String(newAlbumID), newAlbumName);
            showMsg(data.message || '相册创建成功', true);
            if (createTorrentAlbumModal) {
                createTorrentAlbumModal.hide();
            }
            submitFilterWithTorrentAlbumName(newAlbumName);
        }).catch(function (error) {
            showMsg(error && error.message ? error.message : '新增相册失败', false);
            if (createTorrentAlbumConfirmBtn) {
                createTorrentAlbumConfirmBtn.disabled = false;
            }
        });
    }

    if (checkAll) {
        checkAll.addEventListener('change', function () {
            const checked = !!checkAll.checked;
            document.querySelectorAll('.js-torrent-album-item-check').forEach(function (checkbox) {
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
            if (targetAlbumID === torrentAlbumID) {
                showMsg('目标相册不能与当前相册相同', false);
                return;
            }
            batchMoveConfirmBtn.disabled = true;
            const ids = pendingBatchMoveIDs.slice();
            batchMoveTorrentAlbumItems(ids, targetAlbumID);
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
            batchRemoveTorrentAlbumItems(ids);
            if (batchRemoveModal) {
                batchRemoveModal.hide();
            }
        });
    }

    if (openCreateTorrentAlbumBtn) {
        openCreateTorrentAlbumBtn.addEventListener('click', function () {
            if (!createTorrentAlbumModal || !createTorrentAlbumNameEl || !createTorrentAlbumConfirmBtn) {
                showMsg('新增相册弹框未初始化', false);
                return;
            }
            createTorrentAlbumNameEl.value = '';
            createTorrentAlbumConfirmBtn.disabled = false;
            createTorrentAlbumModal.show();
            window.setTimeout(function () {
                createTorrentAlbumNameEl.focus();
            }, 50);
        });
    }

    if (createTorrentAlbumModalEl) {
        createTorrentAlbumModalEl.addEventListener('hidden.bs.modal', function () {
            if (createTorrentAlbumNameEl) {
                createTorrentAlbumNameEl.value = '';
            }
            if (createTorrentAlbumConfirmBtn) {
                createTorrentAlbumConfirmBtn.disabled = false;
            }
        });
    }

    if (createTorrentAlbumConfirmBtn) {
        createTorrentAlbumConfirmBtn.addEventListener('click', function () {
            createTorrentAlbum(createTorrentAlbumNameEl ? createTorrentAlbumNameEl.value : '');
        });
    }

    if (createTorrentAlbumNameEl) {
        createTorrentAlbumNameEl.addEventListener('keydown', function (event) {
            if (event.key !== 'Enter') {
                return;
            }
            event.preventDefault();
            createTorrentAlbum(createTorrentAlbumNameEl.value);
        });
    }

    document.addEventListener('change', function (event) {
        const target = event.target;
        if (target && target.classList && target.classList.contains('js-torrent-album-item-check')) {
            syncSelectionState();
        }
    });

    document.addEventListener('click', function (event) {
        const quickFilterButton = event.target.closest('.js-torrent-album-quick-filter');
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

        const removeButton = event.target.closest('.js-remove-torrent-album-item');
        if (!removeButton) {
            return;
        }
        const itemID = parseItemID(removeButton.getAttribute('data-item-id'));
        if (itemID <= 0) {
            showMsg('缺少相册条目参数，无法移除', false);
            return;
        }
        removeTorrentAlbumItem(itemID, removeButton);
    });

    syncSelectionState();
})();
