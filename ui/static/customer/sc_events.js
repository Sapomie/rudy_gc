(function () {
    const msgArea = document.getElementById('msgArea');
    let msgTimer = null;

    function showMsg(text, ok) {
        if (!msgArea) {
            return;
        }
        clearTimeout(msgTimer);
        msgArea.textContent = text;
        msgArea.className = 'alert ' + (ok ? 'alert-success' : 'alert-danger') + ' fade show small py-2 px-3';
        msgArea.style.display = 'block';
        msgTimer = setTimeout(function () {
            msgArea.style.display = 'none';
        }, 2500);
    }

    function post(url, body) {
        return fetch(url, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: body ? JSON.stringify(body) : null,
        });
    }

    document.querySelectorAll('[data-role="sc-move"]').forEach(function (btn) {
        btn.addEventListener('click', function () {
            const scName = btn.getAttribute('data-sc-name');
            if (!scName) {
                return;
            }

            btn.disabled = true;
            post('/api/triggers/sc/move', {scName: scName})
                .then(function (resp) {
                    if (resp.ok) {
                        showMsg('MoveScFilm 已触发: ' + scName, true);
                        return;
                    }
                    showMsg('触发失败(' + resp.status + ')', false);
                })
                .catch(function (err) {
                    showMsg('异常: ' + err, false);
                })
                .finally(function () {
                    btn.disabled = false;
                });
        });
    });
})();
