(function () {
    const msgArea = document.getElementById('castMsgArea');
    const btn = document.getElementById('btnRebuildActorRank');

    function showMsg(text, ok = true) {
        if (!msgArea) return;
        msgArea.textContent = text;
        msgArea.className = 'alert ' + (ok ? 'alert-success' : 'alert-danger') + ' small py-2 px-3 mb-3';
        msgArea.style.display = 'block';
    }

    if (!btn) return;

    btn.addEventListener('click', function () {
        const actorName = (btn.dataset.actorName || '').trim();
        if (!actorName) {
            showMsg('演员名为空，无法触发', false);
            return;
        }

        fetch('/api/triggers/cast/rebuild-rank', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({actorName}),
        })
            .then((resp) => {
                if (resp.ok) {
                    showMsg('该演员 Rank 回填已触发');
                    return;
                }
                showMsg('触发失败(' + resp.status + ')', false);
            })
            .catch((err) => showMsg('异常：' + err, false));
    });
})();
