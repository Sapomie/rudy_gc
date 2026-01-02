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
    const pickedCard = document.getElementById('pickedCard');
    const pickedGrid = document.getElementById('pickedGrid');
    const pickedCount = document.getElementById('pickedCount');
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

    function escapeHtml(input) {
        return String(input)
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    function renderCard(movie) {
        if (!movie) return '';
        const name = movie.name || '';
        const title = movie.title || '';
        const jacket = movie.jacket_img || '';
        const prefix = movie.prefix || '';
        const releasingDate = movie.releasing_date || '';
        const filmBirthDate = movie.film_birth_date || '';
        const director = movie.director || '';
        const score = Number.isFinite(movie.score) ? movie.score : null;
        const viewersWatched =
            movie.viewers_number_watched !== undefined && movie.viewers_number_watched !== null
                ? movie.viewers_number_watched
                : '';
        const scTimes = Number.isFinite(movie.sc_times) ? movie.sc_times : 0;
        const comeTimes = Number.isFinite(movie.come_times) ? movie.come_times : 0;
        const highestRank = Number.isFinite(movie.highest_rank) ? movie.highest_rank : 0;
        const owned = Number.isFinite(movie.owned) ? movie.owned : 0;
        const busUrl = movie.bus_url || '';
        const searchUrl = movie.search_url || '';
        const javUrl = movie.jav_url || '';
        const videoUrl = movie.video_url || '';
        const genres = Array.isArray(movie.genre) ? movie.genre : [];
        const casts = Array.isArray(movie.cast) ? movie.cast : [];

        const badgeTop =
            (comeTimes > 0 ? '<span class="badge-come">❤ ' + comeTimes + '</span>' : '') +
            (scTimes > 0 ? ' <span class="badge-times">× ' + scTimes + '</span>' : '');

        let badgeBottom = '';
        if (highestRank > 0 || score !== null) {
            badgeBottom =
                '<div class="badge-bottom-right">' +
                (highestRank > 0 ? '<span class="badge-rank">#' + highestRank + '</span>' : '') +
                (score !== null ? '<span class="badge-score">★ ' + score + '</span>' : '') +
                '</div>';
        }

        const viewerInfo =
            score !== null && viewersWatched !== ''
                ? '<small class="text-muted ml-2">(' + escapeHtml(viewersWatched) + ')</small>'
                : '';

        let ownedLine = '';
        if (owned >= 2) {
            let btn = '';
            if (owned === 5) {
                btn = '<a href="' + escapeHtml(videoUrl) + '" class="btn btn-sm btn-info me-2">Play</a>';
            } else if (owned === 4) {
                btn =
                    '<a href="' + escapeHtml(videoUrl) + '" class="btn btn-sm btn-success me-2">Play</a>';
            } else if (owned === 6) {
                btn =
                    '<a href="' +
                    escapeHtml(videoUrl) +
                    '" class="btn btn-sm btn-secondary me-2">ReMoved</a>';
            }
            const filmDate = filmBirthDate ? '<span class="text-muted small">| ' + escapeHtml(filmBirthDate) + '</span>' : '';
            ownedLine =
                '<li class="list-group-item d-flex align-items-center">' + btn + filmDate + '</li>';
        }

        const genreHtml = genres
            .map(function (g) {
                return (
                    '<a href="/cards?gn=' +
                    encodeURIComponent(g) +
                    '" class="badge badge-genre me-1 mb-1">' +
                    escapeHtml(g) +
                    '</a>'
                );
            })
            .join('');

        const castHtml = casts
            .map(function (c) {
                const cname = c && c.name ? c.name : '';
                const cnameShow = c && c.name_show ? c.name_show : cname;
                return (
                    '<a href="/cards?cn=' +
                    encodeURIComponent(cname) +
                    '" class="person-tag mr-2" title="' +
                    escapeHtml(cnameShow) +
                    '">' +
                    escapeHtml(cnameShow) +
                    '</a>'
                );
            })
            .join('');

        const directorHtml = director
            ? '<a href="/cards?dn=' +
              encodeURIComponent(director) +
              '" class="person-tag ml-2" title="' +
              escapeHtml(director) +
              '">' +
              escapeHtml(director) +
              '</a>'
            : '';

        const links =
            (busUrl
                ? '<a href="' + escapeHtml(busUrl) + '" class="me-2" target="_blank" rel="noopener">Bus</a>'
                : '') +
            (searchUrl
                ? '<a href="' +
                  escapeHtml(searchUrl) +
                  '" class="me-2" target="_blank" rel="noopener">Suk</a>'
                : '') +
            (javUrl
                ? '<a href="' + escapeHtml(javUrl) + '" class="me-2" target="_blank" rel="noopener">Lib</a>'
                : '');

        return (
            '<div class="col">' +
            '<article class="card shadow-sm h-100 card-lift">' +
            '<a href="/movie/' +
            encodeURIComponent(name) +
            '" class="d-block" aria-label="查看 ' +
            escapeHtml(name) +
            ' 详情">' +
            '<div class="thumb-2x3">' +
            '<img src="' +
            escapeHtml(jacket) +
            '" alt="' +
            escapeHtml(name) +
            ' 海报" loading="lazy" decoding="async">' +
            '<div class="badge-top-left">' +
            badgeTop +
            '</div>' +
            badgeBottom +
            '</div>' +
            '</a>' +
            '<ul class="list-group list-group-flush">' +
            '<li class="list-group-item">' +
            '<h6 class="mb-1 text-truncate" title="' +
            escapeHtml(title) +
            '">' +
            '<a href="/cards?pn=' +
            encodeURIComponent(prefix) +
            '" class="stretched-link text-reset text-decoration-none">' +
            escapeHtml(name) +
            '</a>' +
            viewerInfo +
            '</h6>' +
            (title ? '<p class="mb-0 text-muted line-clamp-2" title="' + escapeHtml(title) + '">' + escapeHtml(title) + '</p>' : '') +
            '</li>' +
            ownedLine +
            '<li class="list-group-item text-muted">' +
            genreHtml +
            '</li>' +
            '<li class="list-group-item">' +
            castHtml +
            directorHtml +
            '</li>' +
            '<li class="list-group-item d-flex align-items-center small">' +
            links +
            '<span class="ms-auto text-muted">' +
            escapeHtml(releasingDate) +
            '</span>' +
            '</li>' +
            '</ul>' +
            '</article>' +
            '</div>'
        );
    }

    function renderMovies(movies) {
        if (!pickedCard || !pickedGrid || !pickedCount) return;
        const list = Array.isArray(movies) ? movies : [];
        pickedGrid.innerHTML = list.map(renderCard).join('');
        pickedCount.textContent = '共 ' + list.length + ' 部';
        pickedCard.style.display = list.length > 0 ? 'block' : 'none';
        if (list.length > 0) {
            pickedCard.scrollIntoView({behavior: 'smooth', block: 'start'});
        }
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
                renderMovies(data.movies || []);
                showMsg('执行完成，已抽取 ' + picked + ' 部');
            })
            .catch((e) => showMsg('异常：' + e, false));
    });
})();
