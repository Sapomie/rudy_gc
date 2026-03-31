(function () {
    const page = document.getElementById('sehuatangMagnetListPage');
    const msgEl = document.getElementById('sehuatangMagnetListMsg');

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
        button.textContent = favorited ? ('已在' + targetAlbumName) : targetAlbumName;
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
        const favoriteButton = event.target.closest('.js-add-favorite');
        if (!favoriteButton) {
            return;
        }

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
    });
}());
