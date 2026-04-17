(function () {
    const page = document.getElementById('movieAlbumItemsPage');
    if (!page) {
        return;
    }

    const albumID = parseInt(String(page.getAttribute('data-album-id') || '0'), 10);
    const albumName = String(page.getAttribute('data-album-name') || '').trim();
    const msgEl = document.getElementById('movieAlbumItemsMsg');
    const createBtn = document.getElementById('btnOpenCreateMovieAlbumModal');
    const createModalEl = document.getElementById('movieAlbumCreateModal');
    const createNameEl = document.getElementById('movieAlbumCreateName');
    const createConfirmBtn = document.getElementById('movieAlbumCreateConfirmBtn');
    const createModal = (createModalEl && window.bootstrap && window.bootstrap.Modal)
        ? window.bootstrap.Modal.getOrCreateInstance(createModalEl)
        : null;
    const batchExecuteBtn = document.getElementById('btnBatchExecuteMovieAlbumRemove');
    const batchCountEl = document.getElementById('movieAlbumBatchSelectedCount');
    const checkAllEl = document.getElementById('movieAlbumItemCheckAll');
    const executeModalEl = document.getElementById('movieAlbumExecuteRemoveModal');
    const executeCountEl = document.getElementById('movieAlbumExecuteRemoveCount');
    const executeConfirmBtn = document.getElementById('movieAlbumExecuteRemoveConfirmBtn');
    const executeModal = (executeModalEl && window.bootstrap && window.bootstrap.Modal)
        ? window.bootstrap.Modal.getOrCreateInstance(executeModalEl)
        : null;
    const removeDownloadedBtn = document.getElementById('btnRemoveDownloadedMovieAlbumItems');
    const removeDownloadedModalEl = document.getElementById('movieAlbumRemoveDownloadedModal');
    const removeDownloadedConfirmBtn = document.getElementById('movieAlbumRemoveDownloadedConfirmBtn');
    const removeDownloadedCountEl = document.getElementById('movieAlbumRemoveDownloadedCount');
    const removeDownloadedListEl = document.getElementById('movieAlbumRemoveDownloadedList');
    const removeDownloadedEmptyEl = document.getElementById('movieAlbumRemoveDownloadedEmpty');
    const removeDownloadedModal = (removeDownloadedModalEl && window.bootstrap && window.bootstrap.Modal)
        ? window.bootstrap.Modal.getOrCreateInstance(removeDownloadedModalEl)
        : null;
    let removeDownloadedPreviewItems = [];

    function showMsg(text, ok) {
        if (!msgEl) {
            return;
        }
        msgEl.textContent = String(text || '');
        msgEl.className = 'alert small py-2 px-3 mb-3 ' + (ok ? 'alert-success' : 'alert-danger');
    }

    function request(url, options) {
        return fetch(url, options).then(async function (response) {
            const data = await response.json().catch(function () { return {}; });
            if (!response.ok || !data || data.ok === false) {
                throw new Error((data && data.error) || ('请求失败(' + response.status + ')'));
            }
            return data;
        });
    }

    function escapeHtml(value) {
        return String(value || '')
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    function renderRemoveDownloadedPreview(items) {
        removeDownloadedPreviewItems = Array.isArray(items) ? items : [];
        if (removeDownloadedCountEl) {
            removeDownloadedCountEl.textContent = String(removeDownloadedPreviewItems.length);
        }
        if (removeDownloadedListEl) {
            removeDownloadedListEl.innerHTML = removeDownloadedPreviewItems.map(function (item) {
                const name = String((item && item.MovieName) || '').trim();
                const javId = String((item && item.MovieJavID) || '').trim();
                const title = name || javId || '-';
                const meta = javId ? ('<div class="small text-muted"><code>' + escapeHtml(javId) + '</code></div>') : '';
                return '<div class="list-group-item px-0 py-2 border-0 border-bottom">' +
                    '<div class="fw-semibold">' + escapeHtml(title) + '</div>' +
                    meta +
                    '</div>';
            }).join('');
        }
        if (removeDownloadedEmptyEl) {
            if (removeDownloadedPreviewItems.length > 0) {
                removeDownloadedEmptyEl.classList.add('d-none');
            } else {
                removeDownloadedEmptyEl.textContent = '没有可移除的已下载条目。';
                removeDownloadedEmptyEl.classList.remove('d-none');
            }
        }
        if (removeDownloadedConfirmBtn) {
            removeDownloadedConfirmBtn.disabled = removeDownloadedPreviewItems.length === 0;
        }
    }

    function checkedItemIDs() {
        return Array.from(page.querySelectorAll('.js-movie-album-item-check:checked')).map(function (input) {
            return parseInt(String(input.value || '0'), 10);
        }).filter(function (id) {
            return Number.isInteger(id) && id > 0;
        });
    }

    function refreshBatchState() {
        const count = checkedItemIDs().length;
        if (batchCountEl) {
            batchCountEl.textContent = String(count);
        }
        if (batchExecuteBtn) {
            batchExecuteBtn.disabled = !(albumName === '待删除' && count > 0);
        }
        if (checkAllEl) {
            const allChecks = Array.from(page.querySelectorAll('.js-movie-album-item-check'));
            checkAllEl.checked = allChecks.length > 0 && allChecks.every(function (input) { return input.checked; });
        }
    }

    if (createBtn && createModal) {
        createBtn.addEventListener('click', function () {
            if (createNameEl) {
                createNameEl.value = '';
            }
            createModal.show();
        });
    }

    page.addEventListener('click', function (event) {
        const quickBtn = event.target.closest('.js-movie-album-quick-filter');
        if (quickBtn) {
            const nextAlbum = String(quickBtn.getAttribute('data-album-name') || '').trim();
            if (!nextAlbum) {
                return;
            }
            window.location.href = '/movie-albums?album_name=' + encodeURIComponent(nextAlbum);
            return;
        }
    });

    if (createConfirmBtn) {
        createConfirmBtn.addEventListener('click', function () {
            const name = createNameEl ? String(createNameEl.value || '').trim() : '';
            if (!name) {
                showMsg('请输入电影相册名称', false);
                return;
            }
            createConfirmBtn.disabled = true;
            request('/api/movie-albums', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({name: name})
            }).then(function (data) {
                showMsg(data.message || '电影相册创建成功', true);
                if (createModal) {
                    createModal.hide();
                }
                window.location.href = '/movie-albums?album_name=' + encodeURIComponent(String(data.album_name || name));
            }).catch(function (error) {
                showMsg(error && error.message ? error.message : '电影相册创建失败', false);
                createConfirmBtn.disabled = false;
            });
        });
    }

    page.addEventListener('click', function (event) {
        const removeBtn = event.target.closest('.js-remove-movie-album-item');
        if (!removeBtn) {
        } else {
            if (albumID <= 0) {
                showMsg('当前电影相册无效', false);
                return;
            }
            const itemID = parseInt(String(removeBtn.getAttribute('data-item-id') || '0'), 10);
            if (!Number.isInteger(itemID) || itemID <= 0) {
                showMsg('缺少相册条目', false);
                return;
            }
            removeBtn.disabled = true;
            request('/api/movie-albums/' + albumID + '/items/remove', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({item_id: itemID})
            }).then(function (data) {
                const row = page.querySelector('tr[data-item-id="' + itemID + '"]');
                if (row) {
                    row.remove();
                }
                showMsg(data.message || '移除完成', true);
                removeBtn.disabled = false;
                refreshBatchState();
            }).catch(function (error) {
                showMsg(error && error.message ? error.message : '移除失败', false);
                removeBtn.disabled = false;
            });
            return;
        }
    });

    page.addEventListener('change', function (event) {
        if (event.target.closest('.js-movie-album-item-check')) {
            refreshBatchState();
            return;
        }
        if (event.target === checkAllEl) {
            const checked = !!checkAllEl.checked;
            page.querySelectorAll('.js-movie-album-item-check').forEach(function (input) {
                input.checked = checked;
            });
            refreshBatchState();
        }
    });

    if (batchExecuteBtn && executeModal) {
        batchExecuteBtn.addEventListener('click', function () {
            const ids = checkedItemIDs();
            if (ids.length === 0) {
                showMsg('请先选择待删除条目', false);
                return;
            }
            if (executeCountEl) {
                executeCountEl.textContent = String(ids.length);
            }
            executeModal.show();
        });
    }

    if (executeConfirmBtn) {
        executeConfirmBtn.addEventListener('click', function () {
            const ids = checkedItemIDs();
            if (ids.length === 0) {
                showMsg('请先选择待删除条目', false);
                return;
            }
            executeConfirmBtn.disabled = true;
            request('/api/movie-albums/' + albumID + '/execute-remove', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({item_ids: ids})
            }).then(function (data) {
                showMsg(data.message || '统一删除完成', true);
                if (executeModal) {
                    executeModal.hide();
                }
                window.setTimeout(function () {
                    window.location.reload();
                }, 500);
            }).catch(function (error) {
                showMsg(error && error.message ? error.message : '统一删除失败', false);
                executeConfirmBtn.disabled = false;
            });
        });
    }

    if (removeDownloadedBtn && removeDownloadedModal) {
        removeDownloadedBtn.addEventListener('click', function () {
            if (albumID <= 0) {
                showMsg('当前电影相册无效', false);
                return;
            }
            removeDownloadedBtn.disabled = true;
            if (removeDownloadedEmptyEl) {
                removeDownloadedEmptyEl.textContent = '正在加载预览...';
                removeDownloadedEmptyEl.classList.remove('d-none');
            }
            if (removeDownloadedListEl) {
                removeDownloadedListEl.innerHTML = '';
            }
            if (removeDownloadedCountEl) {
                removeDownloadedCountEl.textContent = '0';
            }
            if (removeDownloadedConfirmBtn) {
                removeDownloadedConfirmBtn.disabled = true;
            }
            request('/api/movie-albums/' + albumID + '/remove-downloaded-items-preview', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({})
            }).then(function (data) {
                renderRemoveDownloadedPreview(data.items);
                if (!Array.isArray(data.items) || data.items.length === 0) {
                    showMsg('没有可移除的已下载条目', true);
                    removeDownloadedBtn.disabled = false;
                    return;
                }
                removeDownloadedModal.show();
                removeDownloadedBtn.disabled = false;
            }).catch(function (error) {
                showMsg(error && error.message ? error.message : '加载移除预览失败', false);
                removeDownloadedBtn.disabled = false;
            });
        });
    }

    if (removeDownloadedConfirmBtn) {
        removeDownloadedConfirmBtn.addEventListener('click', function () {
            if (albumID <= 0) {
                showMsg('当前电影相册无效', false);
                return;
            }
            if (removeDownloadedPreviewItems.length === 0) {
                showMsg('没有可移除的已下载条目', false);
                return;
            }
            removeDownloadedConfirmBtn.disabled = true;
            request('/api/movie-albums/' + albumID + '/remove-downloaded-items', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({})
            }).then(function (data) {
                showMsg(data.message || '已移除已下载条目', true);
                if (removeDownloadedModal) {
                    removeDownloadedModal.hide();
                }
                window.setTimeout(function () {
                    window.location.reload();
                }, 500);
            }).catch(function (error) {
                showMsg(error && error.message ? error.message : '移除已下载条目失败', false);
                removeDownloadedConfirmBtn.disabled = false;
            });
        });
    }

    refreshBatchState();
}());
