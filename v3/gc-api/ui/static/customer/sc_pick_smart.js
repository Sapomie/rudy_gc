// /static/customer/sc_pick_smart.js
(function () {
    const msgArea = document.getElementById('msgArea');
    let msgTimer = null;

    function showMsg(text, ok = true) {
        if (!msgArea) return;
        clearTimeout(msgTimer);
        msgArea.textContent = text;
        msgArea.className = 'alert ' + (ok ? 'alert-success' : 'alert-danger') + ' fade show small py-2 px-3';
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

    const form = document.getElementById('formPickSmart');
    const groupWrap = document.getElementById('pickGroups');
    const groupTpl = document.getElementById('pickGroupTpl');
    const btnAdd = document.getElementById('btnAddGroup');
    const btnPickOnly = document.getElementById('btnPickOnly');
    const pickedCard = document.getElementById('pickedCard');
    const pickedGrid = document.getElementById('pickedGrid');
    const pickedCount = document.getElementById('pickedCount');
    const pickedSortBar = document.getElementById('pickedSortBar');
    const pickInfoCard = document.getElementById('pickInfoCard');
    const pickInfoSummary = document.getElementById('pickInfoSummary');
    const pickInfoGrid = document.getElementById('pickInfoGrid');
    const copyCard = document.getElementById('copyCard');
    const copyMeta = document.getElementById('copyMeta');
    const copyProgress = document.getElementById('copyProgress');
    const copyCurrent = document.getElementById('copyCurrent');
    const copyError = document.getElementById('copyError');
    const btnStopCopy = document.getElementById('btnStopCopy');
    if (!form || !groupWrap || !groupTpl || !btnAdd) return;
    const pageSource = (form.getAttribute('data-source') || 'vfilm').toLowerCase() === 'wmedia' ? 'wmedia' : 'vfilm';

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
        const id = 'smart-adv-panel-' + groupSeq;
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

    let pickedMoviesCache = [];
    let pickedTotalSizeGb = '';
    let pickedSortBase = 'film_birth';
    let pickedSortDir = 'desc';

    function numberOrZero(v) {
        return Number.isFinite(v) ? v : 0;
    }

    function parseDateValue(raw) {
        if (!raw) return 0;
        const ts = Date.parse(raw);
        return Number.isFinite(ts) ? ts : 0;
    }

    function rankValue(movie) {
        const v = Number.isFinite(movie && movie.highest_rank) ? movie.highest_rank : 0;
        return v > 0 ? v : Number.MAX_SAFE_INTEGER;
    }

    function compareText(a, b) {
        return String(a || '').localeCompare(String(b || ''), 'zh-Hans-CN', {
            numeric: true,
            sensitivity: 'base',
        });
    }

    function comparePickedOrder(a, b) {
        return numberOrZero(a && a._pickedOrder) - numberOrZero(b && b._pickedOrder);
    }

    function sortPickedMovies(list, sortKey) {
        const sorted = list.slice();
        switch (sortKey) {
            case 'film_birth_desc':
                sorted.sort((a, b) => parseDateValue(b.film_birth_date) - parseDateValue(a.film_birth_date) || comparePickedOrder(a, b));
                break;
            case 'film_birth_asc':
                sorted.sort((a, b) => parseDateValue(a.film_birth_date) - parseDateValue(b.film_birth_date) || comparePickedOrder(a, b));
                break;
            case 'releasing_desc':
                sorted.sort((a, b) => parseDateValue(b.releasing_date) - parseDateValue(a.releasing_date) || comparePickedOrder(a, b));
                break;
            case 'releasing_asc':
                sorted.sort((a, b) => parseDateValue(a.releasing_date) - parseDateValue(b.releasing_date) || comparePickedOrder(a, b));
                break;
            case 'sc_desc':
                sorted.sort((a, b) => numberOrZero(b.sc_times) - numberOrZero(a.sc_times) || comparePickedOrder(a, b));
                break;
            case 'sc_asc':
                sorted.sort((a, b) => numberOrZero(a.sc_times) - numberOrZero(b.sc_times) || comparePickedOrder(a, b));
                break;
            case 'come_desc':
                sorted.sort((a, b) => numberOrZero(b.come_times) - numberOrZero(a.come_times) || comparePickedOrder(a, b));
                break;
            case 'come_asc':
                sorted.sort((a, b) => numberOrZero(a.come_times) - numberOrZero(b.come_times) || comparePickedOrder(a, b));
                break;
            case 'rank_best':
                sorted.sort((a, b) => rankValue(a) - rankValue(b) || comparePickedOrder(a, b));
                break;
            case 'rank_worst':
                sorted.sort((a, b) => rankValue(b) - rankValue(a) || comparePickedOrder(a, b));
                break;
            case 'score_desc':
                sorted.sort((a, b) => numberOrZero(b.score) - numberOrZero(a.score) || comparePickedOrder(a, b));
                break;
            case 'score_asc':
                sorted.sort((a, b) => numberOrZero(a.score) - numberOrZero(b.score) || comparePickedOrder(a, b));
                break;
            case 'name_asc':
                sorted.sort((a, b) => compareText(a.name, b.name) || comparePickedOrder(a, b));
                break;
            default:
                sorted.sort(comparePickedOrder);
                break;
        }
        return sorted;
    }

    function currentSortKey() {
        if (pickedSortBase === 'picked') return 'picked';
        if (pickedSortBase === 'rank') return pickedSortDir === 'desc' ? 'rank_worst' : 'rank_best';
        return pickedSortBase + '_' + pickedSortDir;
    }

    function updatePickedSortButtons() {
        if (!pickedSortBar) return;
        const buttons = pickedSortBar.querySelectorAll('[data-sort-base]');
        buttons.forEach(function (btn) {
            const base = btn.getAttribute('data-sort-base') || '';
            const label = btn.getAttribute('data-label') || btn.textContent.replace(/\s[↑↓]$/, '');
            btn.setAttribute('data-label', label);
            const active = base === pickedSortBase;
            btn.classList.toggle('active', active);
            if (active && base !== 'picked') {
                btn.textContent = label + ' ' + (pickedSortDir === 'asc' ? '↑' : '↓');
            } else {
                btn.textContent = label;
            }
        });
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
        const viewersWatched = movie.viewers_number_watched !== undefined && movie.viewers_number_watched !== null ? movie.viewers_number_watched : '';
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

        const viewerInfo = score !== null && viewersWatched !== ''
            ? '<small class="text-muted ml-2">(' + escapeHtml(viewersWatched) + ')</small>'
            : '';

        let ownedLine = '';
        if (owned >= 2) {
            let btn = '';
            if (owned === 5) {
                btn = '<a href="' + escapeHtml(videoUrl) + '" class="btn btn-sm btn-info me-2">Play</a>';
            } else if (owned === 4) {
                btn = '<a href="' + escapeHtml(videoUrl) + '" class="btn btn-sm btn-success me-2">Play</a>';
            } else if (owned === 6) {
                btn = '<a href="' + escapeHtml(videoUrl) + '" class="btn btn-sm btn-secondary me-2">ReMoved</a>';
            }
            const filmDate = filmBirthDate ? '<span class="text-muted small">| ' + escapeHtml(filmBirthDate) + '</span>' : '';
            ownedLine = '<li class="list-group-item d-flex align-items-center">' + btn + filmDate + '</li>';
        }

        const genreHtml = genres.map(function (g) {
            return '<a href="/cards?gn=' + encodeURIComponent(g) + '" class="badge badge-genre me-1 mb-1">' + escapeHtml(g) + '</a>';
        }).join('');

        const castHtml = casts.map(function (c) {
            const cname = c && c.name ? c.name : '';
            const cnameShow = c && c.name_show ? c.name_show : cname;
            return '<a href="/cards?cn=' + encodeURIComponent(cname) + '" class="person-tag mr-2" title="' + escapeHtml(cnameShow) + '">' + escapeHtml(cnameShow) + '</a>';
        }).join('');

        const directorHtml = director
            ? '<a href="/cards?dn=' + encodeURIComponent(director) + '" class="person-tag ml-2" title="' + escapeHtml(director) + '">' + escapeHtml(director) + '</a>'
            : '';

        const links =
            (busUrl ? '<a href="' + escapeHtml(busUrl) + '" class="me-2" target="_blank" rel="noopener">Bus</a>' : '') +
            (searchUrl ? '<a href="' + escapeHtml(searchUrl) + '" class="me-2" target="_blank" rel="noopener">Suk</a>' : '') +
            (javUrl ? '<a href="' + escapeHtml(javUrl) + '" class="me-2" target="_blank" rel="noopener">Lib</a>' : '');

        return (
            '<div class="col">' +
            '<article class="card shadow-sm h-100 card-lift">' +
            '<a href="/movie/' + encodeURIComponent(name) + '" class="d-block" aria-label="查看 ' + escapeHtml(name) + ' 详情">' +
            '<div class="thumb-2x3">' +
            '<img src="' + escapeHtml(jacket) + '" alt="' + escapeHtml(name) + ' 海报" loading="lazy" decoding="async">' +
            '<div class="badge-top-left">' + badgeTop + '</div>' +
            badgeBottom +
            '</div>' +
            '</a>' +
            '<ul class="list-group list-group-flush">' +
            '<li class="list-group-item">' +
            '<h6 class="mb-1 text-truncate" title="' + escapeHtml(title) + '">' +
            '<a href="/cards?pn=' + encodeURIComponent(prefix) + '" class="stretched-link text-reset text-decoration-none">' + escapeHtml(name) + '</a>' +
            viewerInfo +
            '</h6>' +
            (title ? '<p class="mb-0 text-muted line-clamp-2" title="' + escapeHtml(title) + '">' + escapeHtml(title) + '</p>' : '') +
            '</li>' +
            ownedLine +
            '<li class="list-group-item text-muted">' + genreHtml + '</li>' +
            '<li class="list-group-item">' + castHtml + directorHtml + '</li>' +
            '<li class="list-group-item d-flex align-items-center small">' + links +
            '<span class="ms-auto text-muted">' + escapeHtml(releasingDate) + '</span>' +
            '</li>' +
            '</ul>' +
            '</article>' +
            '</div>'
        );
    }

    function drawPickedMovies(list, totalSizeGb, shouldScroll) {
        if (!pickedCard || !pickedGrid || !pickedCount) return;
        pickedGrid.innerHTML = list.map(renderCard).join('');
        const sizeText = totalSizeGb ? ' · ' + totalSizeGb + ' GB' : '';
        pickedCount.textContent = '共 ' + list.length + ' 部' + sizeText;
        pickedCard.style.display = list.length > 0 ? 'block' : 'none';
        if (shouldScroll && list.length > 0) {
            pickedCard.scrollIntoView({behavior: 'smooth', block: 'start'});
        }
    }

    function applyPickedSort(shouldScroll) {
        drawPickedMovies(sortPickedMovies(pickedMoviesCache, currentSortKey()), pickedTotalSizeGb, shouldScroll);
        updatePickedSortButtons();
    }

    function renderMovies(movies, totalSizeGb) {
        const list = Array.isArray(movies) ? movies : [];
        pickedMoviesCache = list.map(function (movie, idx) {
            return Object.assign({_pickedOrder: idx}, movie);
        });
        pickedTotalSizeGb = totalSizeGb || '';
        applyPickedSort(true);
    }

    function infoNumber(info, key) {
        const value = info && Number(info[key]);
        return Number.isFinite(value) ? value : 0;
    }

    function clearPickInfo() {
        if (pickInfoCard) pickInfoCard.style.display = 'none';
        if (pickInfoSummary) pickInfoSummary.textContent = '';
        if (pickInfoGrid) pickInfoGrid.innerHTML = '';
    }

    function renderPickInfo(info) {
        if (!pickInfoCard || !pickInfoSummary || !pickInfoGrid || !info) {
            clearPickInfo();
            return;
        }
        const totalSizeGb = info.total_size_gb || '';
        const rows = [
            ['原始候选电影', infoNumber(info, 'raw_candidate_movie_count')],
            ['参与计算 Cast', infoNumber(info, 'cast_count')],
            ['阻断后剩余 Cast', infoNumber(info, 'after_block_cast_count')],
            ['阻断后剩余 Movie', infoNumber(info, 'after_block_movie_count')],
            ['总大小(GB)', totalSizeGb || '-'],
            ['Cast LastSc 阻断 Cast', infoNumber(info, 'cast_last_sc_blocked_cast_count')],
            ['LastSc Event 阻断 Cast', infoNumber(info, 'last_sc_event_blocked_cast_count')],
            ['Cast LastSc 影响 Movie', infoNumber(info, 'cast_last_sc_blocked_movie_count')],
            ['LastSc Event 影响 Movie', infoNumber(info, 'last_sc_event_blocked_movie_count')],
            ['演员去重跳过 Movie', infoNumber(info, 'selected_actor_skipped_movie_count')],
        ];
        pickInfoSummary.textContent = '阻断后 Movie ' + infoNumber(info, 'after_block_movie_count') + ' 部';
        pickInfoGrid.innerHTML = rows.map(function (row) {
            return '' +
                '<div class="col-6 col-md-3 col-xl-2">' +
                '<div class="border rounded p-2 h-100">' +
                '<div class="text-muted small">' + escapeHtml(row[0]) + '</div>' +
                '<div class="fw-semibold">' + escapeHtml(row[1]) + '</div>' +
                '</div>' +
                '</div>';
        }).join('');
        pickInfoCard.style.display = 'block';
    }

    let copyTimer = null;
    let copyStatusBootstrapped = false;
    let copyStartedInPage = false;

    function setCopyVisible(visible) {
        if (!copyCard) return;
        copyCard.style.display = visible ? 'block' : 'none';
    }

    function resetCopyStatusUI() {
        setCopyVisible(false);
        if (copyProgress) {
            copyProgress.style.width = '0%';
            copyProgress.textContent = '';
        }
        if (copyMeta) copyMeta.textContent = '';
        if (copyCurrent) copyCurrent.textContent = '';
        if (copyError) {
            copyError.style.display = 'none';
            copyError.textContent = '';
        }
        if (btnStopCopy) btnStopCopy.disabled = true;
    }

    function updateCopyStatus(status) {
        if (!copyCard || !copyMeta || !copyProgress || !copyCurrent || !copyError || !btnStopCopy) return;
        const st = status || {};
        const total = Number.isFinite(st.total) ? st.total : 0;
        const done = Number.isFinite(st.done) ? st.done : 0;
        const running = !!st.running;
        const stopped = !!st.stopped;
        const finishedAt = st.finished_at || 0;
        const current = st.current || '';
        const lastErr = st.last_error || '';

        if (!running && !stopped && !finishedAt && total === 0) {
            setCopyVisible(false);
            return;
        }

        setCopyVisible(true);
        const percent = total > 0 ? Math.min(100, Math.round((done / total) * 100)) : 0;
        copyProgress.style.width = percent + '%';
        copyProgress.textContent = percent > 0 ? percent + '%' : '';

        let stateText = '等待中';
        if (running) stateText = '复制中';
        else if (stopped) stateText = '已停止';
        else if (finishedAt) stateText = '已完成';
        copyMeta.textContent = stateText + ' · ' + done + '/' + total;

        copyCurrent.textContent = current ? '当前：' + current : '';
        if (lastErr) {
            copyError.style.display = 'block';
            copyError.textContent = lastErr;
        } else {
            copyError.style.display = 'none';
            copyError.textContent = '';
        }

        btnStopCopy.disabled = !running;
    }

    function startCopyPolling() {
        if (copyTimer) return;
        copyTimer = setInterval(fetchCopyStatus, 2000);
    }

    function stopCopyPolling() {
        if (!copyTimer) return;
        clearInterval(copyTimer);
        copyTimer = null;
    }

    function fetchCopyStatus() {
        fetch('/api/triggers/sc/copy-status')
            .then((r) => r.json())
            .then((data) => {
                const st = data || {};
                if (!copyStatusBootstrapped) {
                    copyStatusBootstrapped = true;
                    if (!copyStartedInPage && !st.running) {
                        resetCopyStatusUI();
                        stopCopyPolling();
                        return;
                    }
                }

                updateCopyStatus(st);
                if (st.running) startCopyPolling();
                else stopCopyPolling();
            })
            .catch(() => {});
    }

    function buildReqFromGroup(group) {
        const req = {
            Page: 1,
            PageSize: 100000,
        };
        req.MediaOwned = 3;

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

    function buildOptions(n) {
        const castLastScBlockDays = parseNumber(document.getElementById('castLastScBlockDays').value.trim(), false);
        const lastScEventBlockDays = parseNumber(document.getElementById('lastScEventBlockDays').value.trim(), false);
        const rawRank20Min = document.getElementById('rank20Min').value.trim();
        const rawRank100Min = document.getElementById('rank100Min').value.trim();
        const rawRank500Min = document.getElementById('rank500Min').value.trim();
        const castLastScPenaltyAlpha = parseNumber(document.getElementById('castLastScPenaltyAlpha').value.trim(), true);
        const lastScEventPenaltyAlpha = parseNumber(document.getElementById('lastScEventPenaltyAlpha').value.trim(), true);
        const castOwnedScRatioPenaltyAlpha = parseNumber(document.getElementById('castOwnedScRatioPenaltyAlpha').value.trim(), true);
        const movieHasScPenaltyAlpha = parseNumber(document.getElementById('movieHasScPenaltyAlpha').value.trim(), true);

        const rank20Min = rawRank20Min === '' ? 0 : parseNumber(rawRank20Min, false);
        let rank100Min = rawRank100Min === '' ? null : parseNumber(rawRank100Min, false);
        let rank500Min = rawRank500Min === '' ? null : parseNumber(rawRank500Min, false);

        if (rank20Min !== null) {
            if (rank100Min === null) rank100Min = rank20Min;
            if (rank500Min === null) rank500Min = rank100Min;
        }

        if (castLastScBlockDays === null || lastScEventBlockDays === null || rank20Min === null || rank100Min === null || rank500Min === null ||
            castLastScPenaltyAlpha === null || lastScEventPenaltyAlpha === null || castOwnedScRatioPenaltyAlpha === null || movieHasScPenaltyAlpha === null) {
            showMsg('全局抽取策略参数不完整', false);
            return null;
        }
        if (rank20Min > rank100Min || rank100Min > rank500Min || rank500Min > n) {
            showMsg('需要满足 0 <= rank20 <= rank100 <= rank500 <= 抽取数量', false);
            return null;
        }

        return {
            castLastScBlockDays,
            lastScEventBlockDays,
            rank20Min,
            rank100Min,
            rank500Min,
            castLastScPenaltyAlpha,
            lastScEventPenaltyAlpha,
            castOwnedScRatioPenaltyAlpha,
            movieHasScPenaltyAlpha,
        };
    }

    function buildPayload() {
        if (!form.checkValidity()) {
            form.classList.add('was-validated');
            return null;
        }

        const rawN = document.getElementById('pickCount').value.trim();
        const n = parseInt(rawN, 10);
        if (!Number.isFinite(n) || n <= 0) {
            showMsg('请输入大于 0 的数量', false);
            return null;
        }

        const options = buildOptions(n);
        if (!options) return null;

        const groups = groupWrap.querySelectorAll('.pick-group');
        if (groups.length === 0) {
            showMsg('请添加至少一组条件', false);
            return null;
        }

        const reqs = [];
        for (const g of groups) {
            const weightEl = g.querySelector('[data-role="weight"]');
            const w = weightEl ? parseInt(weightEl.value.trim(), 10) : 0;
            if (!Number.isFinite(w) || w <= 0) {
                showMsg('权重需为正数', false);
                return null;
            }
            reqs.push({
                weight: w,
                req: buildReqFromGroup(g),
            });
        }

        return {pickN: n, source: pageSource, options, reqs};
    }

    function runPick(url, actionLabel) {
        const payload = buildPayload();
        if (!payload) return;

        clearPickInfo();
        post(url, payload)
            .then(async (r) => {
                if (!r.ok) {
                    const data = await r.json().catch(() => ({}));
                    const msg = data.error ? data.error : '触发失败(' + r.status + ')';
                    showMsg(msg, false);
                    return;
                }
                const data = await r.json().catch(() => ({}));
                const picked = data.picked || 0;
                renderPickInfo(data.pick_info || null);
                renderMovies(data.movies || [], data.total_size_gb || '');
                if (data.copy_status) {
                    copyStatusBootstrapped = true;
                    copyStartedInPage = !!(data.copy_started || data.copy_status.running);
                    updateCopyStatus(data.copy_status);
                    if (data.copy_status.running) startCopyPolling();
                }
                if (data.copy_started === false && data.copy_status && data.copy_status.running) {
                    showMsg('Copy 任务正在执行，本次不再启动', false);
                }
                showMsg((actionLabel || '执行完成') + '，已抽取 ' + picked + ' 部');
            })
            .catch((e) => showMsg('异常：' + e, false));
    }

    form.addEventListener('submit', function (e) {
        e.preventDefault();
        runPick('/api/triggers/sc/pick-smart-copy', 'Smart Pick + Copy 完成');
    });

    if (btnPickOnly) {
        btnPickOnly.addEventListener('click', function () {
            runPick('/api/triggers/sc/pick-smart-only', 'Smart Pick 完成');
        });
    }

    if (btnStopCopy) {
        btnStopCopy.addEventListener('click', function () {
            post('/api/triggers/sc/copy-stop', {})
                .then((r) => r.json().catch(() => ({})))
                .then((data) => {
                    copyStatusBootstrapped = true;
                    updateCopyStatus(data.status || {});
                    showMsg('已发送停止指令');
                    stopCopyPolling();
                })
                .catch((e) => showMsg('停止失败：' + e, false));
        });
    }

    if (pickedSortBar) {
        pickedSortBar.addEventListener('click', function (e) {
            const btn = e.target.closest('[data-sort-base]');
            if (!btn) return;
            const base = btn.getAttribute('data-sort-base') || 'picked';
            const defaultDir = btn.getAttribute('data-default-dir') || 'desc';
            if (base === pickedSortBase) {
                if (base !== 'picked') {
                    pickedSortDir = pickedSortDir === 'asc' ? 'desc' : 'asc';
                }
            } else {
                pickedSortBase = base;
                pickedSortDir = defaultDir;
            }
            applyPickedSort(false);
        });
    }

    fetchCopyStatus();
})();
