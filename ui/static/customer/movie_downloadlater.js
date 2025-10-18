// movie_downloadlater.js
document.addEventListener('click', async function (e) {
    const btn = e.target.closest('#btn-downloadlater');
    if (!btn) return;

    const name = btn.dataset.name;
    const current = btn.dataset.needdownload === '2'; // 当前状态：是否已添加
    const method = current ? 'DELETE' : 'POST';
    const url = `/api/movie/${encodeURIComponent(name)}/downloadlater`;

    try {
        const res = await fetch(url, {method});
        const data = await res.json();
        if (!res.ok || !data.ok) throw new Error(data.error || '操作失败');

        // ✅ 2 = 稍后下载, 1 = 未添加
        const isAdded = data.needDownload === 2;
        btn.dataset.needdownload = isAdded ? '2' : '1';

        // UI 切换
        btn.classList.toggle('btn-warning', isAdded);
        btn.classList.toggle('btn-outline-warning', !isAdded);
        const icon = btn.querySelector('i');
        icon.classList.toggle('mdi-download-outline', !isAdded);
        icon.classList.toggle('mdi-download', isAdded);
        const span = btn.querySelector('span');
        span.textContent = isAdded ? '已添加' : '稍后下载';
        btn.title = isAdded ? '取消稍后下载' : '添加稍后下载';
    } catch (err) {
        alert(err.message || '网络错误，请稍后重试');
    }
});