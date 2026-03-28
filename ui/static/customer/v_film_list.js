(function () {
    const msgArea = document.getElementById('filmMsgArea');

    function showMsg(text, ok) {
        if (!msgArea) return;
        msgArea.textContent = text;
        msgArea.className = 'alert mb-3 ' + (ok ? 'alert-success' : 'alert-danger');
    }

    function setText(row, role, value) {
        if (!row) return;
        const el = row.querySelector('[data-role="' + role + '"]');
        if (!el) return;
        el.textContent = value;
    }

    function updateRow(row, result) {
        if (!row || !result) return;
        setText(row, 'film-height', result.height > 0 ? String(result.height) : '-');
        setText(row, 'film-duration', result.duration_minutes > 0 ? String(result.duration_minutes) : '-');
        setText(row, 'film-bitrate', result.bit_rate > 0 ? String(result.bit_rate) : '-');
        setText(row, 'film-frame', result.frame_average || '-');
    }

    async function handleProbe(btn) {
        const filmId = btn.getAttribute('data-film-id');
        if (!filmId) return;

        const row = btn.closest('tr');
        const oldText = btn.textContent;
        btn.disabled = true;
        btn.textContent = 'scanning...';

        try {
            const resp = await fetch('/api/films/' + encodeURIComponent(filmId) + '/probe-meta', {
                method: 'POST',
            });
            const data = await resp.json();
            if (!resp.ok || !data || !data.ok) {
                throw new Error((data && data.error) || 'scan_meta failed');
            }

            updateRow(row, data.result);
            showMsg('scan_meta 完成', true);
        } catch (err) {
            showMsg(err && err.message ? err.message : 'scan_meta 失败', false);
        } finally {
            btn.disabled = false;
            btn.textContent = oldText;
        }
    }

    document.addEventListener('click', function (event) {
        const btn = event.target.closest('.js-film-probe');
        if (!btn) return;
        event.preventDefault();
        handleProbe(btn);
    });
})();
