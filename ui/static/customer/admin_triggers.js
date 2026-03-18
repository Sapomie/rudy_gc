// /static/customer/admin_triggers.js
(function () {
    const msgArea = document.getElementById('msgArea');
    let msgTimer = null;

    function showMsg(text, ok = true) {
        if (!msgArea) return;
        clearTimeout(msgTimer);
        msgArea.textContent = text;
        msgArea.className =
            'alert ' + (ok ? 'alert-success' : 'alert-danger') + ' fade show small py-2 px-3';
        msgArea.style.display = 'block';
        msgTimer = setTimeout(() => (msgArea.style.display = 'none'), 2500);
    }

    function post(url, body) {
        return fetch(url, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: body ? JSON.stringify(body) : null,
        });
    }

    // 通用按钮绑定
    function bindBtn(id, url, okMsg) {
        const el = document.getElementById(id);
        if (!el) return;
        el.onclick = function () {
            post(url)
                .then((r) => (r.ok ? showMsg(okMsg) : showMsg('触发失败(' + r.status + ')', false)))
                .catch((e) => showMsg('异常：' + e, false));
        };
    }

    // 爬虫按钮
    bindBtn('btnDailyBest', '/api/triggers/daily-best', 'DailyBest 已触发');
    bindBtn('btnDailyBestSync', '/api/triggers/daily-best-sync', 'DailyBest 同步 已触发');
    bindBtn('btnRebuildCastRank', '/api/triggers/rebuild-cast-rank', 'Rank 回填 已触发');
    bindBtn('btnSeeds', '/api/triggers/seeds', 'Seeds 已触发');

    // 影片按钮
    bindBtn('btnFilmRename', '/api/triggers/film/rename', '影片重命名 已触发');
    bindBtn('btnFilmProcess', '/api/triggers/film/process', '影片处理 已触发');
    bindBtn('btnScRebuildStats', '/api/triggers/sc/rebuild-stats', 'SC 回填 已触发');

    // === 刷新最久未更新详情 ===
    const formRefreshOldest = document.getElementById('formRefreshOldest');
    if (formRefreshOldest) {
        formRefreshOldest.addEventListener('submit', function (e) {
            e.preventDefault();

            if (!formRefreshOldest.checkValidity()) {
                formRefreshOldest.classList.add('was-validated');
                return;
            }

            const raw = document.getElementById('refreshNumber').value.trim();
            const n = parseInt(raw, 10);
            if (!Number.isFinite(n) || n <= 0) {
                showMsg('请输入大于 0 的数字', false);
                return;
            }

            post('/api/triggers/refresh-oldest-detail', {number: n})
                .then((r) =>
                    r.ok ? showMsg('刷新最久详情 已触发') : showMsg('触发失败(' + r.status + ')', false),
                )
                .catch((e) => showMsg('异常：' + e, false));
        });
    }

    // === 按 Seed 名称触发 ===
    const formSeedByName = document.getElementById('formSeedByName');
    if (formSeedByName) {
        formSeedByName.addEventListener('submit', function (e) {
            e.preventDefault();
            if (!formSeedByName.checkValidity()) {
                formSeedByName.classList.add('was-validated');
                return;
            }
            const name = document.getElementById('seedByName').value.trim();
            post('/api/triggers/seed-by-name', {name})
                .then((r) =>
                    r.ok ? showMsg('SeedByName 已触发') : showMsg('触发失败(' + r.status + ')', false),
                )
                .catch((e) => showMsg('异常：' + e, false));
        });
    }

})();
