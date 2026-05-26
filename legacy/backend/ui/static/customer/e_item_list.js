(function () {
    const tableBodyEl = document.getElementById('itemTableBody');
    const totalCountEl = document.getElementById('itemTotalCount');

    const singleModalEl = document.getElementById('itemStatusModal');
    const singleForm = document.getElementById('itemStatusForm');
    const singleItemIdInput = document.getElementById('itemStatusItemId');
    const singleMetaEl = document.getElementById('itemStatusMeta');
    const singleMsgEl = document.getElementById('itemStatusMsg');
    const singleSubmitBtn = document.getElementById('itemStatusSubmit');
    const singleHasDetailEl = document.getElementById('itemStatusHasDetail');
    const singleHasDownloadCoverEl = document.getElementById('itemStatusHasDownloadCover');
    const singleHasChineseEl = document.getElementById('itemStatusHasChinese');
    const singleDetailNeedScanEl = document.getElementById('itemStatusDetailNeedScan');

    const batchModalEl = document.getElementById('itemBatchStatusModal');
    const batchForm = document.getElementById('itemBatchStatusForm');
    const batchMetaEl = document.getElementById('itemBatchStatusMeta');
    const batchMsgEl = document.getElementById('itemBatchStatusMsg');
    const batchSubmitBtn = document.getElementById('itemBatchStatusSubmit');
    const batchHasDetailEl = document.getElementById('itemBatchHasDetail');
    const batchHasDownloadCoverEl = document.getElementById('itemBatchHasDownloadCover');
    const batchHasChineseEl = document.getElementById('itemBatchHasChinese');
    const batchDetailNeedScanEl = document.getElementById('itemBatchDetailNeedScan');
    const batchEditBtn = document.getElementById('btnBatchEditStatus');
    const selectedCountEl = document.getElementById('itemBatchSelectedCount');
    const checkAllEl = document.getElementById('itemCheckAll');

    const deleteModalEl = document.getElementById('itemDeleteModal');
    const deleteMetaEl = document.getElementById('itemDeleteMeta');
    const deleteMsgEl = document.getElementById('itemDeleteMsg');
    const deleteItemIdInput = document.getElementById('itemDeleteItemId');
    const deleteSubmitBtn = document.getElementById('itemDeleteSubmit');

    if (!window.bootstrap || !singleModalEl || !singleForm) {
        return;
    }

    const singleModal = bootstrap.Modal.getOrCreateInstance(singleModalEl);
    const batchModal = batchModalEl ? bootstrap.Modal.getOrCreateInstance(batchModalEl) : null;
    const deleteModal = deleteModalEl ? bootstrap.Modal.getOrCreateInstance(deleteModalEl) : null;

    function showMsg(el, text, ok) {
        if (!el) return;
        el.textContent = text;
        el.className = 'alert small py-2 px-3 mt-3 mb-0 ' + (ok ? 'alert-success' : 'alert-danger');
    }

    function hideMsg(el) {
        if (!el) return;
        el.textContent = '';
        el.className = 'alert small py-2 px-3 mt-3 mb-0 d-none';
    }

    function setSubmitting(button, submitting, idleText, submittingText) {
        if (!button) return;
        button.disabled = submitting;
        button.textContent = submitting ? submittingText : idleText;
    }

    function setSelectValue(selectEl, value) {
        if (!selectEl) return;
        selectEl.value = value || selectEl.options[0].value;
    }

    function getItemRows() {
        return Array.from(document.querySelectorAll('tr[data-item-id]'));
    }

    function getItemCheckboxes() {
        return Array.from(document.querySelectorAll('.js-item-check'));
    }

    function getSelectedIds() {
        return getItemCheckboxes()
            .filter(function (checkbox) {
                return checkbox.checked;
            })
            .map(function (checkbox) {
                return Number(checkbox.value);
            })
            .filter(function (id) {
                return Number.isInteger(id) && id > 0;
            });
    }

    function updateButtonDataset(button, item) {
        if (!button || !item) return;
        button.dataset.hasDetail = item.has_detail_value || '';
        button.dataset.hasDownloadCover = item.has_download_cover_value || '';
        button.dataset.hasChinese = item.has_chinese_value || '';
        button.dataset.detailNeedScan = item.detail_need_scan_value || '';
    }

    function updateRow(item) {
        if (!item) return;
        const row = document.querySelector('tr[data-item-id="' + item.id + '"]');
        if (!row) return;

        const hasDetailCell = row.querySelector('.js-item-has-detail');
        const hasDownloadCoverCell = row.querySelector('.js-item-has-download-cover');
        const hasChineseCell = row.querySelector('.js-item-has-chinese');
        const detailNeedScanCell = row.querySelector('.js-item-detail-need-scan');
        const editBtn = row.querySelector('.js-item-status-edit');

        if (hasDetailCell) hasDetailCell.textContent = item.has_detail_text || '-';
        if (hasDownloadCoverCell) hasDownloadCoverCell.textContent = item.has_download_cover_text || '-';
        if (hasChineseCell) hasChineseCell.textContent = item.has_chinese_text || '-';
        if (detailNeedScanCell) detailNeedScanCell.textContent = item.detail_need_scan_text || '-';
        updateButtonDataset(editBtn, item);
    }

    function updateTotalCount(delta) {
        if (!totalCountEl) return;
        const current = Number(totalCountEl.textContent || '0');
        if (!Number.isFinite(current)) return;
        const next = current + delta;
        totalCountEl.textContent = String(next > 0 ? next : 0);
    }

    function renderEmptyStateIfNeeded() {
        if (!tableBodyEl) return;
        if (getItemRows().length > 0) return;
        tableBodyEl.innerHTML = [
            '<tr class="js-item-empty-row">',
            '    <td colspan="10">',
            '        <div class="sc-empty-state">当前筛选下没有 e_item 数据。</div>',
            '    </td>',
            '</tr>'
        ].join('');
    }

    function removeEmptyState() {
        const emptyRow = tableBodyEl ? tableBodyEl.querySelector('.js-item-empty-row') : null;
        if (emptyRow) {
            emptyRow.remove();
        }
    }

    function syncSelectionState() {
        const checkboxes = getItemCheckboxes();
        const checkedCount = checkboxes.filter(function (checkbox) {
            return checkbox.checked;
        }).length;

        if (selectedCountEl) {
            selectedCountEl.textContent = String(checkedCount);
        }
        if (batchEditBtn) {
            batchEditBtn.disabled = checkedCount === 0;
        }
        if (checkAllEl) {
            checkAllEl.checked = checkboxes.length > 0 && checkedCount === checkboxes.length;
            checkAllEl.indeterminate = checkedCount > 0 && checkedCount < checkboxes.length;
        }

        checkboxes.forEach(function (checkbox) {
            const row = checkbox.closest('tr');
            if (!row) return;
            row.classList.toggle('table-active', checkbox.checked);
        });
    }

    function clearSelection() {
        getItemCheckboxes().forEach(function (checkbox) {
            checkbox.checked = false;
        });
        if (checkAllEl) {
            checkAllEl.checked = false;
            checkAllEl.indeterminate = false;
        }
        syncSelectionState();
    }

    function buildBatchPayload(ids) {
        const payload = {ids: ids};
        let changedFieldCount = 0;

        if (batchHasDetailEl && batchHasDetailEl.value) {
            payload.hasDetail = batchHasDetailEl.value;
            changedFieldCount += 1;
        }
        if (batchHasDownloadCoverEl && batchHasDownloadCoverEl.value) {
            payload.hasDownloadCover = batchHasDownloadCoverEl.value;
            changedFieldCount += 1;
        }
        if (batchHasChineseEl && batchHasChineseEl.value) {
            payload.hasChinese = batchHasChineseEl.value;
            changedFieldCount += 1;
        }
        if (batchDetailNeedScanEl && batchDetailNeedScanEl.value) {
            payload.detailNeedScan = batchDetailNeedScanEl.value;
            changedFieldCount += 1;
        }

        return {payload: payload, changedFieldCount: changedFieldCount};
    }

    function removeRow(itemId) {
        const row = document.querySelector('tr[data-item-id="' + itemId + '"]');
        if (!row) return;
        row.remove();
        updateTotalCount(-1);
        syncSelectionState();
        renderEmptyStateIfNeeded();
    }

    document.addEventListener('change', function (event) {
        const checkbox = event.target.closest('.js-item-check');
        if (checkbox) {
            syncSelectionState();
            return;
        }

        if (checkAllEl && event.target === checkAllEl) {
            getItemCheckboxes().forEach(function (itemCheckbox) {
                itemCheckbox.checked = checkAllEl.checked;
            });
            syncSelectionState();
        }
    });

    document.addEventListener('click', function (event) {
        const editButton = event.target.closest('.js-item-status-edit');
        if (editButton) {
            singleItemIdInput.value = editButton.dataset.itemId || '';
            if (singleMetaEl) {
                singleMetaEl.textContent = (editButton.dataset.itemJavId || '-') + ' / ' + (editButton.dataset.itemName || '-');
            }
            setSelectValue(singleHasDetailEl, editButton.dataset.hasDetail);
            setSelectValue(singleHasDownloadCoverEl, editButton.dataset.hasDownloadCover);
            setSelectValue(singleHasChineseEl, editButton.dataset.hasChinese);
            setSelectValue(singleDetailNeedScanEl, editButton.dataset.detailNeedScan);
            hideMsg(singleMsgEl);
            singleModal.show();
            return;
        }

        const deleteButton = event.target.closest('.js-item-delete');
        if (deleteButton) {
            if (deleteItemIdInput) {
                deleteItemIdInput.value = deleteButton.dataset.itemId || '';
            }
            if (deleteMetaEl) {
                deleteMetaEl.textContent = (deleteButton.dataset.itemJavId || '-') + ' / ' + (deleteButton.dataset.itemName || '-');
            }
            hideMsg(deleteMsgEl);
            if (deleteModal) {
                deleteModal.show();
            }
            return;
        }

        if (batchEditBtn && event.target === batchEditBtn) {
            const ids = getSelectedIds();
            if (ids.length === 0) {
                return;
            }
            if (batchMetaEl) {
                batchMetaEl.textContent = '已选 ' + ids.length + ' 项';
            }
            setSelectValue(batchHasDetailEl, '');
            setSelectValue(batchHasDownloadCoverEl, '');
            setSelectValue(batchHasChineseEl, '');
            setSelectValue(batchDetailNeedScanEl, '');
            hideMsg(batchMsgEl);
            if (batchModal) {
                batchModal.show();
            }
        }
    });

    singleModalEl.addEventListener('hidden.bs.modal', function () {
        hideMsg(singleMsgEl);
        setSubmitting(singleSubmitBtn, false, '保存', '保存中...');
    });

    if (batchModalEl) {
        batchModalEl.addEventListener('hidden.bs.modal', function () {
            hideMsg(batchMsgEl);
            setSubmitting(batchSubmitBtn, false, '批量保存', '保存中...');
        });
    }

    if (deleteModalEl) {
        deleteModalEl.addEventListener('hidden.bs.modal', function () {
            hideMsg(deleteMsgEl);
            setSubmitting(deleteSubmitBtn, false, '确认删除', '删除中...');
        });
    }

    singleForm.addEventListener('submit', async function (event) {
        event.preventDefault();
        hideMsg(singleMsgEl);

        const itemId = (singleItemIdInput.value || '').trim();
        if (!itemId) {
            showMsg(singleMsgEl, '缺少条目 id', false);
            return;
        }

        setSubmitting(singleSubmitBtn, true, '保存', '保存中...');
        try {
            const response = await fetch('/api/e-items/' + encodeURIComponent(itemId) + '/status', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    hasDetail: singleHasDetailEl.value,
                    hasDownloadCover: singleHasDownloadCoverEl.value,
                    hasChinese: singleHasChineseEl.value,
                    detailNeedScan: singleDetailNeedScanEl.value
                })
            });
            const payload = await response.json().catch(function () {
                return {};
            });

            if (!response.ok || !payload.ok) {
                showMsg(singleMsgEl, payload.error || ('保存失败(' + response.status + ')'), false);
                return;
            }

            updateRow(payload.item);
            showMsg(singleMsgEl, '状态已更新', true);
            setTimeout(function () {
                singleModal.hide();
            }, 250);
        } catch (error) {
            showMsg(singleMsgEl, '异常：' + error, false);
        } finally {
            setSubmitting(singleSubmitBtn, false, '保存', '保存中...');
        }
    });

    if (batchForm) {
        batchForm.addEventListener('submit', async function (event) {
            event.preventDefault();
            hideMsg(batchMsgEl);

            const ids = getSelectedIds();
            if (ids.length === 0) {
                showMsg(batchMsgEl, '请先选择条目', false);
                return;
            }

            const built = buildBatchPayload(ids);
            if (built.changedFieldCount === 0) {
                showMsg(batchMsgEl, '至少选择一个要修改的状态', false);
                return;
            }

            setSubmitting(batchSubmitBtn, true, '批量保存', '保存中...');
            try {
                const response = await fetch('/api/e-items/status/batch', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(built.payload)
                });
                const payload = await response.json().catch(function () {
                    return {};
                });

                if (!response.ok || !payload.ok) {
                    showMsg(batchMsgEl, payload.error || ('保存失败(' + response.status + ')'), false);
                    return;
                }

                if (Array.isArray(payload.items)) {
                    payload.items.forEach(updateRow);
                }
                clearSelection();
                showMsg(batchMsgEl, '批量状态已更新', true);
                setTimeout(function () {
                    if (batchModal) {
                        batchModal.hide();
                    }
                }, 250);
            } catch (error) {
                showMsg(batchMsgEl, '异常：' + error, false);
            } finally {
                setSubmitting(batchSubmitBtn, false, '批量保存', '保存中...');
            }
        });
    }

    if (deleteSubmitBtn) {
        deleteSubmitBtn.addEventListener('click', async function () {
            hideMsg(deleteMsgEl);

            const itemId = (deleteItemIdInput && deleteItemIdInput.value ? deleteItemIdInput.value : '').trim();
            if (!itemId) {
                showMsg(deleteMsgEl, '缺少条目 id', false);
                return;
            }

            setSubmitting(deleteSubmitBtn, true, '确认删除', '删除中...');
            try {
                const response = await fetch('/api/e-items/' + encodeURIComponent(itemId), {
                    method: 'DELETE'
                });
                const payload = await response.json().catch(function () {
                    return {};
                });

                if (!response.ok || !payload.ok) {
                    showMsg(deleteMsgEl, payload.error || ('删除失败(' + response.status + ')'), false);
                    return;
                }

                removeEmptyState();
                removeRow(Number(payload.id || itemId));
                showMsg(deleteMsgEl, '电影已删除', true);
                setTimeout(function () {
                    if (deleteModal) {
                        deleteModal.hide();
                    }
                }, 250);
            } catch (error) {
                showMsg(deleteMsgEl, '异常：' + error, false);
            } finally {
                setSubmitting(deleteSubmitBtn, false, '确认删除', '删除中...');
            }
        });
    }

    syncSelectionState();
})();
