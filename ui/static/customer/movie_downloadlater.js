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
});