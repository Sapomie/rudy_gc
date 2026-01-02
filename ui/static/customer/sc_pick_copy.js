// /static/customer/sc_pick_copy.js
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
        msgTimer = setTimeout(() => (msgArea.style.display = 'none'), 3000);
    }

    function post(url, body) {
        return fetch(url, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(body),
        });
    }

    const form = document.getElementById('formPickCopy');
    const groupWrap = document.getElementById('pickGroups');
    const groupTpl = document.getElementById('pickGroupTpl');
    const btnAdd = document.getElementById('btnAddGroup');
    if (!form || !groupWrap || !groupTpl || !btnAdd) return;

    function bindRemove(btn) {
        btn.addEventListener('click', function () {
            const group = btn.closest('.pick-group');
            if (!group) return;
            if (groupWrap.children.length <= 1) {
                showMsg('至少保留一组条件', false);
                return;
            }
            group.remove();
        });
    }

    let groupSeq = 0;

    function setupAdvanced(group) {
        const toggle = group.querySelector('[data-role="adv-toggle"]');
        const panel = group.querySelector('[data-role="adv-panel"]');
        if (!toggle || !panel) return;

        groupSeq += 1;
        const id = 'adv-panel-' + groupSeq;
        panel.id = id;
        toggle.setAttribute('data-bs-toggle', 'collapse');
        toggle.setAttribute('data-bs-target', '#' + id);
        toggle.setAttribute('aria-controls', id);
        toggle.setAttribute('aria-expanded', 'false');
    }

    function addGroup() {
        const node = document.importNode(groupTpl.content, true);
        const group = node.querySelector('.pick-group');
        const btnRemove = node.querySelector('.btn-remove-group');
        if (btnRemove) bindRemove(btnRemove);
        if (group) setupAdvanced(group);
        groupWrap.appendChild(group);
    }

    btnAdd.addEventListener('click', function () {
        addGroup();
    });

    addGroup();

    function parseNumber(raw, isFloat) {
        const v = isFloat ? parseFloat(raw) : parseInt(raw, 10);
        return Number.isFinite(v) ? v : null;
    }

    function buildReqFromGroup(group) {
        const req = {
            Owned: 3,
            Page: 1,
            PageSize: 1000,
        };

        const inputs = group.querySelectorAll('[data-req-field]');
        inputs.forEach((input) => {
            const key = input.getAttribute('data-req-field');
            const type = input.getAttribute('data-type') || 'string';
            const raw = input.value.trim();
            if (!key) return;

            if (type === 'int') {
                const v = raw === '' ? 0 : parseNumber(raw, false);
                req[key] = v === null ? 0 : v;
                return;
            }
            if (type === 'int-opt') {
                if (raw === '') return;
                const v = parseNumber(raw, false);
                if (v !== null) req[key] = v;
                return;
            }
            if (type === 'float') {
                const v = raw === '' ? 0 : parseNumber(raw, true);
                req[key] = v === null ? 0 : v;
                return;
            }
            if (type === 'float-opt') {
                if (raw === '') return;
                const v = parseNumber(raw, true);
                if (v !== null) req[key] = v;
                return;
            }
            if (raw !== '') req[key] = raw;
        });

        return req;
    }

    form.addEventListener('submit', function (e) {
        e.preventDefault();

        if (!form.checkValidity()) {
            form.classList.add('was-validated');
            return;
        }

        const rawN = document.getElementById('pickCount').value.trim();
        const n = parseInt(rawN, 10);
        if (!Number.isFinite(n) || n <= 0) {
            showMsg('请输入大于 0 的数量', false);
            return;
        }

        const groups = groupWrap.querySelectorAll('.pick-group');
        if (groups.length === 0) {
            showMsg('请添加至少一组条件', false);
            return;
        }

        const reqs = [];
        for (const g of groups) {
            const weightEl = g.querySelector('[data-role="weight"]');
            const w = weightEl ? parseInt(weightEl.value.trim(), 10) : 0;
            if (!Number.isFinite(w) || w <= 0) {
                showMsg('权重需为正数', false);
                return;
            }
            reqs.push({
                weight: w,
                req: buildReqFromGroup(g),
            });
        }

        post('/api/triggers/sc/pick-copy', {pickN: n, reqs})
            .then(async (r) => {
                if (!r.ok) {
                    const data = await r.json().catch(() => ({}));
                    const msg = data.error ? data.error : '触发失败(' + r.status + ')';
                    showMsg(msg, false);
                    return;
                }
                const data = await r.json().catch(() => ({}));
                const picked = data.picked || 0;
                showMsg('执行完成，已抽取 ' + picked + ' 部');
            })
            .catch((e) => showMsg('异常：' + e, false));
    });
})();
