(function () {
    const msgArea = document.getElementById('msgArea');
    let msgTimer = null;

    function showMsg(text, ok = true) {
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
            headers: { 'Content-Type': 'application/json' },
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

    bindBtn('btnDailyBest', '/api/triggers/daily-best', 'DailyBest 已触发');
    bindBtn('btnSeeds', '/api/triggers/seeds', 'Seeds 已触发');
    bindBtn('btnFilmRename', '/api/triggers/film/rename', '影片重命名 已触发');
    bindBtn('btnFilmProcess', '/api/triggers/film/process', '影片处理 已触发');

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
            post('/api/triggers/seed-by-name', { name })
                .then((r) =>
                    r.ok ? showMsg('SeedByName 已触发') : showMsg('触发失败(' + r.status + ')', false)
                )
                .catch((e) => showMsg('异常：' + e, false));
        });
    }

    // === SC Move ===
    const formMove = document.getElementById('formScMove');
    if (formMove) {
        formMove.addEventListener('submit', function (e) {
            e.preventDefault();
            if (!formMove.checkValidity()) {
                formMove.classList.add('was-validated');
                return;
            }
            const scName = document.getElementById('scMoveName').value.trim();
            post('/api/triggers/sc/move', { scName })
                .then((r) =>
                    r.ok ? showMsg('MoveScFilm 已触发') : showMsg('触发失败(' + r.status + ')', false)
                )
                .catch((e) => showMsg('异常：' + e, false));
        });
    }

    // === SC Add ===
    const formAdd = document.getElementById('formScAdd');
    if (formAdd) {
        formAdd.addEventListener('submit', function (e) {
            e.preventDefault();
            if (!formAdd.checkValidity()) {
                formAdd.classList.add('was-validated');
                return;
            }
            const dir = document.getElementById('scAddDir').value.trim();
            post('/api/triggers/sc/add', { dir })
                .then((r) =>
                    r.ok ? showMsg('AddSc 已触发') : showMsg('触发失败(' + r.status + ')', false)
                )
                .catch((e) => showMsg('异常：' + e, false));
        });
    }
})();