(function () {
    const page = document.getElementById('fetchSiteListPage');
    if (!page) {
        return;
    }

    const movieJavID = String(page.getAttribute('data-movie-jav-id') || '').trim();
    const openBtn = document.getElementById('btnOpenMovieAlbumModal');
    const modalEl = document.getElementById('movieAlbumModal');
    const selectEl = document.getElementById('movieAlbumSelect');
    const newNameEl = document.getElementById('movieAlbumNewName');
    const addBtn = document.getElementById('btnAddMovieToAlbum');
    const createBtn = document.getElementById('btnCreateMovieAlbumInline');
    const msgEl = document.getElementById('movieAlbumMsg');
    const modal = (modalEl && window.bootstrap && window.bootstrap.Modal)
        ? window.bootstrap.Modal.getOrCreateInstance(modalEl)
        : null;

    function showMsg(text, ok) {
        if (!msgEl) {
            return;
        }
        msgEl.textContent = String(text || '');
        msgEl.className = 'alert mb-3 ' + (ok ? 'alert-success' : 'alert-danger');
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

    function appendAlbumOption(albumName, selected) {
        if (!selectEl) {
            return;
        }
        const name = String(albumName || '').trim();
        if (!name) {
            return;
        }
        const exists = Array.from(selectEl.options).some(function (option) {
            return String(option.value || '').trim() === name;
        });
        if (!exists) {
            const option = document.createElement('option');
            option.value = name;
            option.textContent = name;
            selectEl.appendChild(option);
        }
        if (selected) {
            selectEl.value = name;
        }
    }

    if (openBtn && modal) {
        openBtn.addEventListener('click', function () {
            if (msgEl) {
                msgEl.className = 'alert d-none mb-3';
                msgEl.textContent = '';
            }
            modal.show();
        });
    }

    if (createBtn) {
        createBtn.addEventListener('click', function () {
            const name = newNameEl ? String(newNameEl.value || '').trim() : '';
            if (!name) {
                showMsg('请输入电影相册名称', false);
                return;
            }
            createBtn.disabled = true;
            request('/api/movie-albums', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({name: name})
            }).then(function (data) {
                appendAlbumOption(data.album_name || name, true);
                if (newNameEl) {
                    newNameEl.value = '';
                }
                showMsg(data.message || '电影相册创建成功', true);
                createBtn.disabled = false;
            }).catch(function (error) {
                showMsg(error && error.message ? error.message : '电影相册创建失败', false);
                createBtn.disabled = false;
            });
        });
    }

    if (addBtn) {
        addBtn.addEventListener('click', function () {
            const albumName = selectEl ? String(selectEl.value || '').trim() : '';
            if (!movieJavID) {
                showMsg('缺少影片编号', false);
                return;
            }
            if (!albumName) {
                showMsg('请选择电影相册', false);
                return;
            }
            addBtn.disabled = true;
            request('/api/movie/' + encodeURIComponent(movieJavID) + '/movie-album-item', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({movie_jav_id: movieJavID, album_name: albumName})
            }).then(function (data) {
                showMsg(data.message || '已加入电影相册', true);
                window.setTimeout(function () {
                    window.location.reload();
                }, 500);
            }).catch(function (error) {
                showMsg(error && error.message ? error.message : '加入电影相册失败', false);
                addBtn.disabled = false;
            });
        });
    }
}());
