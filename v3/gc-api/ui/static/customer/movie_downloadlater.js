document.addEventListener('click', async function (e) {
    // === 稍后下载按钮 ===
    const btnLater = e.target.closest('#btn-downloadlater');
    if (btnLater) {
        const name = btnLater.dataset.name;
        const current = btnLater.dataset.needdownload === '2'; // 当前状态：是否已添加
        const method = current ? 'DELETE' : 'POST';
        const url = `/api/movie/${encodeURIComponent(name)}/downloadlater`;

        try {
            const res = await fetch(url, {method});
            const data = await res.json();
            if (!res.ok || !data.ok) throw new Error(data.error || '操作失败');

            const isAdded = data.needDownload === 2;
            btnLater.dataset.needdownload = isAdded ? '2' : '1';

            // UI 切换
            btnLater.classList.toggle('btn-warning', isAdded);
            btnLater.classList.toggle('btn-outline-warning', !isAdded);
            const icon = btnLater.querySelector('i');
            icon.classList.toggle('mdi-download-outline', !isAdded);
            icon.classList.toggle('mdi-download', isAdded);
            const span = btnLater.querySelector('span');
            span.textContent = isAdded ? '已添加' : '稍后下载';
            btnLater.title = isAdded ? '取消稍后下载' : '添加稍后下载';
        } catch (err) {
            alert(err.message || '网络错误，请稍后重试');
        }
        return; // 防止继续匹配下面按钮
    }

    // === 下载封面按钮 ===
    const btnCover = e.target.closest('#btnDownloadCover');
    if (btnCover) {
        const javId = btnCover.dataset.jav;
        if (!javId) return;
        btnCover.disabled = true;
        btnCover.classList.add('disabled');

        try {
            const res = await fetch(`/api/movie/${encodeURIComponent(javId)}/download-cover`, {method: 'POST'});
            if (res.ok) {
                alert('封面下载完成');
            } else {
                const data = await res.json().catch(() => ({}));
                alert('下载失败：' + (data.error || res.status));
            }
        } catch (e2) {
            alert('请求异常：' + e2);
        } finally {
            btnCover.disabled = false;
            btnCover.classList.remove('disabled');
        }
    }

    // === WMedia 移动到 Removed ===
    const btnMoveWMedia = e.target.closest('.js-move-wmedia-removed');
    if (btnMoveWMedia) {
        const javId = (btnMoveWMedia.dataset.jav || '').trim();
        if (!javId) return;
        const modalEl = document.getElementById('moveWMediaRemovedModal');
        if (modalEl) {
            modalEl.dataset.jav = javId;
            const msgEl = modalEl.querySelector('#moveWMediaRemovedMsg');
            if (msgEl) {
                msgEl.style.display = 'none';
                msgEl.textContent = '';
                msgEl.className = 'small mt-2';
            }
            if (window.bootstrap && window.bootstrap.Modal) {
                window.bootstrap.Modal.getOrCreateInstance(modalEl).show();
            }
        }
    }
});

document.addEventListener('submit', async function (e) {
    const formAddCast = e.target.closest('#formAddMovieCast');
    if (!formAddCast) return;
    e.preventDefault();

    const javId = (formAddCast.dataset.jav || '').trim();
    const input = formAddCast.querySelector('#movieCastName');
    const msg = formAddCast.querySelector('#movieCastMsg');
    const btn = formAddCast.querySelector('button[type="submit"]');
    const name = input ? input.value.trim() : '';
    const modalEl = document.getElementById('addMovieCastModal');

    function showMsg(text, ok) {
        if (!msg) return;
        msg.textContent = text;
        msg.className = (ok ? 'text-success' : 'text-danger') + ' small';
        msg.style.display = 'block';
    }

    if (!javId) {
        showMsg('缺少影片参数', false);
        return;
    }
    if (!name) {
        showMsg('演员名不能为空', false);
        return;
    }

    if (btn) btn.disabled = true;
    try {
        const res = await fetch(`/api/movie/${encodeURIComponent(javId)}/add-cast`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({name}),
        });
        const data = await res.json().catch(() => ({}));
        if (!res.ok || !data.ok) {
            throw new Error(data.error || '添加失败');
        }
        showMsg('添加成功，正在刷新...', true);
        window.setTimeout(function () {
            if (modalEl && window.bootstrap && window.bootstrap.Modal) {
                const modal = window.bootstrap.Modal.getOrCreateInstance(modalEl);
                modal.hide();
            }
            window.location.reload();
        }, 300);
    } catch (err) {
        showMsg(err.message || '网络错误，请稍后重试', false);
    } finally {
        if (btn) btn.disabled = false;
    }
});

document.addEventListener('click', async function (e) {
    const btnDo = e.target.closest('#btnMoveWMediaRemovedDo');
    if (!btnDo) return;
    const modalEl = document.getElementById('moveWMediaRemovedModal');
    const javId = modalEl ? (modalEl.dataset.jav || '').trim() : '';
    if (!javId) return;

    btnDo.disabled = true;
    try {
        const res = await fetch(`/api/movie/${encodeURIComponent(javId)}/move-wmedia-removed`, {method: 'POST'});
        const data = await res.json().catch(() => ({}));
        if (!res.ok || !data.ok) {
            throw new Error(data.error || '移动失败');
        }
        const msgEl = modalEl ? modalEl.querySelector('#moveWMediaRemovedMsg') : null;
        if (msgEl) {
            msgEl.textContent = data.message || '已加入待删除相册';
            msgEl.className = 'small mt-2 text-success';
            msgEl.style.display = 'block';
        }
        window.setTimeout(function () {
            if (modalEl && window.bootstrap && window.bootstrap.Modal) {
                window.bootstrap.Modal.getOrCreateInstance(modalEl).hide();
            }
            window.location.reload();
        }, 400);
    } catch (err) {
        const msgEl = modalEl ? modalEl.querySelector('#moveWMediaRemovedMsg') : null;
        if (msgEl) {
            msgEl.textContent = err.message || '网络错误，请稍后重试';
            msgEl.className = 'small mt-2 text-danger';
            msgEl.style.display = 'block';
        }
    } finally {
        btnDo.disabled = false;
    }
});

document.addEventListener('shown.bs.modal', function (e) {
    const modal = e.target.closest('#addMovieCastModal');
    if (!modal) return;
    const input = modal.querySelector('#movieCastName');
    const msg = modal.querySelector('#movieCastMsg');
    if (msg) {
        msg.style.display = 'none';
        msg.textContent = '';
    }
    if (input) {
        input.value = '';
        input.focus();
    }
});
