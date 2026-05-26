(function () {
    const page = document.getElementById('sehuatangMagnetListPage');
    const msgEl = document.getElementById('sehuatangMagnetListMsg');
    const filterForm = document.getElementById('sehuatangMagnetFilterForm');
    const tagInput = document.getElementById('sehuatangMagnetTagInput');
    const otherAlbumModalEl = document.getElementById('sehuatangOtherAlbumModal');
    const otherAlbumModal = otherAlbumModalEl && window.bootstrap && window.bootstrap.Modal
        ? new window.bootstrap.Modal(otherAlbumModalEl)
        : null;
    const dateModalEl = document.getElementById('sehuatangDateModal');
    const dateForm = document.getElementById('sehuatangDateForm');
    const dateModal = dateModalEl && window.bootstrap && window.bootstrap.Modal
        ? new window.bootstrap.Modal(dateModalEl)
        : null;
    const dateTitleEl = document.getElementById('sehuatangDateModalTitle');
    const dateItemKeyEl = document.getElementById('sehuatangDateItemKey');
    const dateInputEl = document.getElementById('sehuatangDateInput');
    const dateTodayBtn = document.getElementById('sehuatangDateTodayBtn');
    const dateMsgEl = document.getElementById('sehuatangDateMsg');
    const dateSubmitBtn = document.getElementById('sehuatangDateSubmit');
    const editShtTimeBtn = document.getElementById('btnEditShtTime');
    let pendingOtherAlbumTarget = null;

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

    function setButtonState(button, loading) {
        if (!button) {
            return;
        }
        button.disabled = loading;
    }

    function showInlineMsg(el, text, ok) {
        if (!el) {
            return;
        }
        el.textContent = String(text || '');
        el.className = 'alert mb-0 ' + (ok ? 'alert-success' : 'alert-danger');
    }

    function hideInlineMsg(el) {
        if (!el) {
            return;
        }
        el.textContent = '';
        el.className = 'alert d-none mb-0';
    }

    function normalizeDateInputValue(raw) {
        const digits = String(raw || '').replace(/\D/g, '').slice(0, 8);
        if (digits.length <= 4) {
            return digits;
        }
        if (digits.length <= 6) {
            return digits.slice(0, 4) + '-' + digits.slice(4);
        }
        return digits.slice(0, 4) + '-' + digits.slice(4, 6) + '-' + digits.slice(6);
    }

    function isValidDateValue(raw) {
        return /^[0-9]{4}-[0-9]{2}-[0-9]{2}$/.test(String(raw || '').trim());
    }

    function todayDateString() {
        const now = new Date();
        const year = String(now.getFullYear());
        const month = String(now.getMonth() + 1).padStart(2, '0');
        const day = String(now.getDate()).padStart(2, '0');
        return year + '-' + month + '-' + day;
    }

    function resolveAlbumName(button, fallback) {
        const raw = button ? String(button.getAttribute('data-album-name') || '').trim() : '';
        if (raw) {
            return raw;
        }
        const base = String(fallback || '').trim();
        return base || '下载中';
    }

    function applyFavoriteButtonState(button, favorited, albumName) {
        if (!button) {
            return;
        }
        const targetAlbumName = resolveAlbumName(button, albumName);
        button.setAttribute('data-favorited', favorited ? '1' : '0');
        button.setAttribute('data-album-name', targetAlbumName);
        button.classList.remove('btn-warning', 'btn-outline-warning', 'btn-info', 'btn-outline-info');

        const pendingAlbum = targetAlbumName === '待下载';
        if (pendingAlbum) {
            button.classList.add(favorited ? 'btn-info' : 'btn-outline-info');
        } else {
            button.classList.add(favorited ? 'btn-warning' : 'btn-outline-warning');
        }
        button.textContent = targetAlbumName;
        button.title = favorited ? ('移出' + targetAlbumName) : targetAlbumName;
    }

    function readMovieTarget(button) {
        if (!button) {
            return {movieRouteKey: '', movieJavID: '', movieName: ''};
        }
        const row = button.closest('tr[data-movie-route-key]');
        if (!row) {
            return {movieRouteKey: '', movieJavID: '', movieName: ''};
        }
        return {
            movieRouteKey: String(row.getAttribute('data-movie-route-key') || '').trim(),
            movieJavID: String(row.getAttribute('data-movie-jav-id') || '').trim(),
            movieName: String(row.getAttribute('data-movie-name') || '').trim()
        };
    }

    function switchAlbumFavorite(movieRouteKey, movieJavID, sourceType, sourceRowID, currentFavorited, button, albumName) {
        const targetAlbumName = resolveAlbumName(button, albumName);
        const method = currentFavorited ? 'DELETE' : 'POST';
        setButtonState(button, true);
        request('/api/movie/' + encodeURIComponent(movieRouteKey) + '/album-item', {
            method: method,
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                movie_jav_id: movieJavID,
                source_type: sourceType,
                source_row_id: sourceRowID,
                album_name: targetAlbumName
            })
        }).then(function (data) {
            const favorited = data && typeof data.favorited === 'boolean' ? data.favorited : !currentFavorited;
            const responseAlbumName = data && data.album_name ? String(data.album_name).trim() : targetAlbumName;
            applyFavoriteButtonState(button, favorited, responseAlbumName);
            const message = data && data.message ? String(data.message) : '操作成功';
            showMsg(message, true);
            setButtonState(button, false);
        }).catch(function (error) {
            showMsg(error && error.message ? error.message : (targetAlbumName + '操作失败'), false);
            setButtonState(button, false);
        });
    }

    function addToOtherAlbum(movieRouteKey, movieJavID, sourceType, sourceRowID, button, albumName) {
        const targetAlbumName = String(albumName || '').trim();
        if (!targetAlbumName) {
            showMsg('请选择其他相册', false);
            return;
        }
        setButtonState(button, true);
        request('/api/movie/' + encodeURIComponent(movieRouteKey) + '/album-item', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                movie_jav_id: movieJavID,
                source_type: sourceType,
                source_row_id: sourceRowID,
                album_name: targetAlbumName
            })
        }).then(function (data) {
            const message = data && data.message ? String(data.message) : ('已加入相册：' + targetAlbumName);
            showMsg(message, true);
            setButtonState(button, false);
        }).catch(function (error) {
            showMsg(error && error.message ? error.message : ('加入相册失败：' + targetAlbumName), false);
            setButtonState(button, false);
        });
    }

    function openOtherAlbumModal(button) {
        if (!button || !otherAlbumModal) {
            return;
        }
        const target = readMovieTarget(button);
        const movieRouteKey = target.movieRouteKey;
        const movieJavID = target.movieJavID;
        const sourceType = String(button.getAttribute('data-source-type') || '').trim();
        const sourceRowID = parseInt(String(button.getAttribute('data-source-row-id') || '0'), 10);
        if (!movieRouteKey || !sourceType || !Number.isInteger(sourceRowID) || sourceRowID <= 0) {
            showMsg('缺少其他相册参数，无法执行', false);
            return;
        }
        pendingOtherAlbumTarget = {
            movieRouteKey: movieRouteKey,
            movieJavID: movieJavID,
            sourceType: sourceType,
            sourceRowID: sourceRowID,
            button: button
        };
        otherAlbumModal.show();
    }

    function openDateModal(button) {
        if (!button || !dateModal || !dateInputEl || !dateItemKeyEl) {
            return;
        }
        const label = String(button.getAttribute('data-item-label') || '').trim() || '日期';
        const itemKey = String(button.getAttribute('data-item-key') || '').trim();
        const itemValue = String(button.getAttribute('data-item-value') || '').trim();
        if (!itemKey) {
            showMsg('缺少日期 key，无法执行', false);
            return;
        }
        if (dateTitleEl) {
            dateTitleEl.textContent = '修改' + label;
        }
        dateItemKeyEl.value = itemKey;
        dateInputEl.value = itemValue;
        hideInlineMsg(dateMsgEl);
        dateModal.show();
    }

    if (dateInputEl) {
        dateInputEl.addEventListener('input', function () {
            const normalized = normalizeDateInputValue(dateInputEl.value);
            if (dateInputEl.value !== normalized) {
                dateInputEl.value = normalized;
            }
        });
    }

    if (dateTodayBtn && dateInputEl) {
        dateTodayBtn.addEventListener('click', function () {
            dateInputEl.value = todayDateString();
            hideInlineMsg(dateMsgEl);
            dateInputEl.focus();
        });
    }

    function applyDayGroupColors() {
        const rows = Array.from(document.querySelectorAll('tr.js-sehuatang-row[data-post-day]'));
        let currentDay = '';
        let currentGroup = 'a';
        rows.forEach(function (row) {
            const postDay = String(row.getAttribute('data-post-day') || '').trim();
            row.classList.remove('fetch-site-day-group-a', 'fetch-site-day-group-b');
            if (!postDay) {
                return;
            }
            if (currentDay === '') {
                currentDay = postDay;
            } else if (postDay !== currentDay) {
                currentDay = postDay;
                currentGroup = currentGroup === 'a' ? 'b' : 'a';
            }
            row.classList.add(currentGroup === 'a' ? 'fetch-site-day-group-a' : 'fetch-site-day-group-b');
        });
    }

    applyDayGroupColors();

    document.addEventListener('click', function (event) {
        const tagFilterLink = event.target.closest('.js-sehuatang-tag-filter');
        if (tagFilterLink) {
            event.preventDefault();
            const targetTag = String(tagFilterLink.getAttribute('data-tag') || '').trim();
            if (!targetTag || !filterForm || !tagInput) {
                return;
            }
            tagInput.value = targetTag;
            filterForm.submit();
            return;
        }

        const favoriteButton = event.target.closest('.js-add-favorite');
        if (favoriteButton) {
            const target = readMovieTarget(favoriteButton);
            const movieRouteKey = target.movieRouteKey;
            const movieJavID = target.movieJavID;
            const sourceType = String(favoriteButton.getAttribute('data-source-type') || '').trim();
            const sourceRowID = parseInt(String(favoriteButton.getAttribute('data-source-row-id') || '0'), 10);
            const currentFavorited = String(favoriteButton.getAttribute('data-favorited') || '').trim() === '1';
            const albumName = resolveAlbumName(favoriteButton);
            if (!movieRouteKey || !sourceType || !Number.isInteger(sourceRowID) || sourceRowID <= 0) {
                showMsg('缺少' + albumName + '参数，无法执行', false);
                return;
            }
            switchAlbumFavorite(movieRouteKey, movieJavID, sourceType, sourceRowID, currentFavorited, favoriteButton, albumName);
            return;
        }

        const openOtherAlbumButton = event.target.closest('.js-open-other-album-modal');
        if (openOtherAlbumButton) {
            openOtherAlbumModal(openOtherAlbumButton);
            return;
        }

        const openDateButton = event.target.closest('#btnEditShtTime');
        if (openDateButton) {
            openDateModal(openDateButton);
            return;
        }

        const otherAlbumOption = event.target.closest('.js-other-album-option');
        if (!otherAlbumOption) {
            return;
        }
        const albumName = String(otherAlbumOption.getAttribute('data-album-name') || '').trim();
        if (!pendingOtherAlbumTarget) {
            showMsg('缺少其他相册目标，无法执行', false);
            return;
        }
        const target = pendingOtherAlbumTarget;
        addToOtherAlbum(target.movieRouteKey, target.movieJavID, target.sourceType, target.sourceRowID, target.button, albumName);
        if (otherAlbumModal) {
            otherAlbumModal.hide();
        }
    });

    if (dateForm && dateInputEl && dateItemKeyEl) {
        dateForm.addEventListener('submit', function (event) {
            event.preventDefault();
            const itemKey = String(dateItemKeyEl.value || '').trim();
            const itemValue = String(dateInputEl.value || '').trim();
            if (!itemKey) {
                showInlineMsg(dateMsgEl, '缺少日期 key', false);
                return;
            }
            if (!itemValue) {
                showInlineMsg(dateMsgEl, '请选择日期', false);
                return;
            }
            if (!isValidDateValue(itemValue)) {
                showInlineMsg(dateMsgEl, '日期格式必须为 YYYY-MM-DD', false);
                return;
            }

            setButtonState(dateSubmitBtn, true);
            request('/api/w-kv/date', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    item_key: itemKey,
                    item_value: itemValue
                })
            }).then(function (data) {
                const savedValue = data && data.item_value ? String(data.item_value).trim() : itemValue;
                if (editShtTimeBtn) {
                    editShtTimeBtn.setAttribute('data-item-value', savedValue);
                    editShtTimeBtn.textContent = 'SHT时间：' + savedValue;
                }
                showMsg('SHT时间已更新', true);
                hideInlineMsg(dateMsgEl);
                setButtonState(dateSubmitBtn, false);
                if (dateModal) {
                    dateModal.hide();
                }
            }).catch(function (error) {
                showInlineMsg(dateMsgEl, error && error.message ? error.message : '保存失败', false);
                setButtonState(dateSubmitBtn, false);
            });
        });
    }
}());
