(function () {
    const msgArea = document.getElementById('msgArea');
    let msgTimer = null;

    const GLIST_IS_NOT_SC = 1;
    const GLIST_IS_SC = 2;

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

    const btnScAddPreview = document.getElementById('btnScAddPreview');
    const scPreviewPanel = document.getElementById('scPreviewPanel');
    const formScAddConfirm = document.getElementById('formScAddConfirm');
    const scPreviewMsg = document.getElementById('scAddPreviewMsg');
    const scPreviewName = document.getElementById('scPreviewName');
    const scPreviewImage = document.getElementById('scPreviewImage');
    const scPreviewMovieCount = document.getElementById('scPreviewMovieCount');
    const scPreviewSelectedMovieCount = document.getElementById('scPreviewSelectedMovieCount');
    const scPreviewMovieTableBody = document.getElementById('scPreviewMovieTableBody');
    const scPreviewComeMovie = document.getElementById('scPreviewComeMovie');
    const scPreviewMovieCast = document.getElementById('scPreviewMovieCast');
    const scPreviewKind = document.getElementById('scPreviewKind');
    const scPreviewDuration = document.getElementById('scPreviewDuration');
    const scPreviewFg = document.getElementById('scPreviewFg');
    const scPreviewVessel = document.getElementById('scPreviewVessel');
    const scPreviewRemarks = document.getElementById('scPreviewRemarks');
    let scPreviewData = null;

    function setScPreviewMsg(text, ok) {
        if (!scPreviewMsg) return;
        if (!text) {
            scPreviewMsg.className = 'small mb-3';
            scPreviewMsg.textContent = '';
            return;
        }
        scPreviewMsg.className = 'alert ' + (ok ? 'alert-success' : 'alert-danger') + ' small py-2 px-3 mb-3';
        scPreviewMsg.textContent = text;
    }

    function normalizePreviewMovies(movies) {
        if (!Array.isArray(movies)) {
            return [];
        }
        return movies.map((movie) => ({
            movieName: movie && movie.movieName ? movie.movieName : '',
            movieJavId: movie && movie.movieJavId ? movie.movieJavId : '',
            movieHref: movie && movie.movieHref ? movie.movieHref : '',
            jacketImg: movie && movie.jacketImg ? movie.jacketImg : '',
            casts: Array.isArray(movie && movie.casts) ? movie.casts.slice() : [],
            isSc: GLIST_IS_SC,
        }));
    }

    function getSelectedScMovies() {
        if (!scPreviewData || !Array.isArray(scPreviewData.movies)) {
            return [];
        }
        return scPreviewData.movies.filter((movie) => movie && movie.isSc === GLIST_IS_SC);
    }

    function findMovieByJavId(movieJavId) {
        if (!scPreviewData || !Array.isArray(scPreviewData.movies)) {
            return null;
        }
        return scPreviewData.movies.find((item) => item.movieJavId === movieJavId) || null;
    }

    function buildCastOptions(movieJavId) {
        if (!scPreviewMovieCast) return;
        scPreviewMovieCast.innerHTML = '';
        const movie = findMovieByJavId(movieJavId);
        if (!movie || !Array.isArray(movie.casts) || movie.casts.length === 0) {
            const option = document.createElement('option');
            option.value = '';
            option.textContent = '可为空';
            scPreviewMovieCast.appendChild(option);
            return;
        }
        const emptyOption = document.createElement('option');
        emptyOption.value = '';
        emptyOption.textContent = '可为空';
        scPreviewMovieCast.appendChild(emptyOption);
        movie.casts.forEach((name) => {
            const option = document.createElement('option');
            option.value = name;
            option.textContent = name;
            scPreviewMovieCast.appendChild(option);
        });
    }

    function syncSelectedMovieCount() {
        if (!scPreviewSelectedMovieCount) return;
        scPreviewSelectedMovieCount.value = String(getSelectedScMovies().length);
    }

    function renderMovieTable() {
        if (!scPreviewMovieTableBody) return;
        scPreviewMovieTableBody.innerHTML = '';

        if (!scPreviewData || !Array.isArray(scPreviewData.movies) || scPreviewData.movies.length === 0) {
            const empty = document.createElement('div');
            empty.className = 'text-muted';
            empty.textContent = '暂无预览数据';
            scPreviewMovieTableBody.appendChild(empty);
            return;
        }

        scPreviewData.movies.forEach((movie, idx) => {
            const item = document.createElement('div');

            const card = document.createElement('div');
            card.className = 'sc-preview-movie-card';

            const posterWrap = document.createElement(movie.movieHref ? 'a' : 'div');
            if (movie.movieHref) {
                posterWrap.href = movie.movieHref;
            }

            if (movie.jacketImg) {
                const poster = document.createElement('img');
                poster.className = 'sc-preview-movie-poster';
                poster.src = movie.jacketImg;
                poster.alt = (movie.movieName || '') + ' 海报';
                poster.loading = 'lazy';
                poster.decoding = 'async';
                posterWrap.appendChild(poster);
            } else {
                const posterEmpty = document.createElement('div');
                posterEmpty.className = 'sc-preview-movie-poster-empty';
                posterWrap.appendChild(posterEmpty);
            }

            const title = document.createElement(movie.movieHref ? 'a' : 'div');
            title.className = 'sc-preview-movie-link fw-semibold';
            if (movie.movieHref) {
                title.href = movie.movieHref;
            }
            title.textContent = movie.movieName || '';

            const radios = document.createElement('div');
            radios.className = 'sc-preview-is-sc-buttons';

            const radioName = 'sc-preview-is-sc-' + idx;
            radios.appendChild(buildIsScRadio(radioName, idx, GLIST_IS_NOT_SC, movie.isSc === GLIST_IS_NOT_SC));
            radios.appendChild(buildIsScRadio(radioName, idx, GLIST_IS_SC, movie.isSc === GLIST_IS_SC));

            card.appendChild(posterWrap);
            card.appendChild(title);
            card.appendChild(radios);
            item.appendChild(card);
            scPreviewMovieTableBody.appendChild(item);
        });
    }

    function buildIsScRadio(groupName, idx, value, checked) {
        const wrap = document.createElement('div');

        const input = document.createElement('input');
        input.type = 'radio';
        input.className = 'btn-check sc-preview-is-sc-input-' + String(value);
        input.name = groupName;
        input.value = String(value);
        input.checked = checked;
        input.autocomplete = 'off';
        input.id = groupName + '-' + String(value);
        input.addEventListener('change', function () {
            const targetIndex = parseInt(String(idx), 10);
            if (!Number.isFinite(targetIndex) || !scPreviewData || !Array.isArray(scPreviewData.movies)) {
                return;
            }
            const currentMovie = scPreviewData.movies[targetIndex];
            if (!currentMovie) {
                return;
            }
            currentMovie.isSc = value === GLIST_IS_SC ? GLIST_IS_SC : GLIST_IS_NOT_SC;
            rebuildComeMovieOptions();
        });

        const label = document.createElement('label');
        label.className = 'btn btn-outline-secondary sc-preview-is-sc-btn sc-preview-is-sc-btn-' + String(value);
        label.setAttribute('for', input.id);
        label.textContent = value === GLIST_IS_SC ? '计入SC' : '不计入';

        wrap.appendChild(input);
        wrap.appendChild(label);
        return wrap;
    }

    function rebuildComeMovieOptions() {
        if (!scPreviewComeMovie) return;
        const selectedScMovies = getSelectedScMovies();
        const previousValue = (scPreviewComeMovie.value || '').trim();
        scPreviewComeMovie.innerHTML = '';

        const emptyOption = document.createElement('option');
        emptyOption.value = '';
        emptyOption.textContent = '可为空';
        scPreviewComeMovie.appendChild(emptyOption);

        selectedScMovies.forEach((movie) => {
            const option = document.createElement('option');
            option.value = movie.movieJavId;
            option.textContent = movie.movieName;
            if (movie.movieJavId === previousValue) {
                option.selected = true;
            }
            scPreviewComeMovie.appendChild(option);
        });

        if (scPreviewComeMovie.options.length > 0) {
            if (![...scPreviewComeMovie.options].some((option) => option.selected)) {
                scPreviewComeMovie.options[0].selected = true;
            }
            buildCastOptions(scPreviewComeMovie.value);
        } else {
            buildCastOptions('');
        }
        syncSelectedMovieCount();
    }

    function fillScPreview(data) {
        scPreviewData = {
            scName: data.scName || '',
            imageFound: !!data.imageFound,
            imageName: data.imageName || '',
            movieCount: data.movieCount || 0,
            movies: normalizePreviewMovies(data.movies),
        };
        if (scPreviewName) scPreviewName.value = scPreviewData.scName;
        if (scPreviewImage) scPreviewImage.value = scPreviewData.imageFound ? (scPreviewData.imageName || '已发现图片') : '无图片';
        if (scPreviewMovieCount) scPreviewMovieCount.value = String(scPreviewData.movieCount || 0);
        if (scPreviewKind) scPreviewKind.value = '';
        if (scPreviewDuration) scPreviewDuration.value = '';
        if (scPreviewFg) scPreviewFg.value = '';
        if (scPreviewVessel) scPreviewVessel.value = '';
        if (scPreviewRemarks) scPreviewRemarks.value = '';
        setScPreviewMsg('', true);

        renderMovieTable();
        rebuildComeMovieOptions();
    }

    function buildPayload() {
        const payload = {
            comeMovieJavId: (scPreviewComeMovie && scPreviewComeMovie.value || '').trim(),
            movieCast: (scPreviewMovieCast && scPreviewMovieCast.value || '').trim(),
            kind: (scPreviewKind && scPreviewKind.value || '').trim(),
            fg: (scPreviewFg && scPreviewFg.value || '').trim(),
            vessel: (scPreviewVessel && scPreviewVessel.value || '').trim(),
            remarks: (scPreviewRemarks && scPreviewRemarks.value || '').trim(),
            movies: (scPreviewData && Array.isArray(scPreviewData.movies) ? scPreviewData.movies : []).map((movie) => ({
                movieJavId: movie.movieJavId,
                isSc: movie.isSc === GLIST_IS_SC ? GLIST_IS_SC : GLIST_IS_NOT_SC,
            })),
        };

        const rawDurationMinutes = (scPreviewDuration && scPreviewDuration.value || '').trim();
        if (rawDurationMinutes === '') {
            payload.duration = 0;
        } else {
            const parsedDurationMinutes = parseInt(rawDurationMinutes, 10);
            if (!Number.isFinite(parsedDurationMinutes)) {
                throw new Error('时长(分钟)必须是数字');
            }
            payload.duration = parsedDurationMinutes;
        }
        return payload;
    }

    if (scPreviewComeMovie) {
        scPreviewComeMovie.addEventListener('change', function () {
            buildCastOptions(scPreviewComeMovie.value);
        });
    }

    if (btnScAddPreview) {
        btnScAddPreview.addEventListener('click', function () {
            fetch('/api/triggers/sc/add-preview', {method: 'POST'})
                .then(async (r) => {
                    const data = await r.json().catch(() => ({}));
                    if (!r.ok) {
                        throw new Error(data.error || ('触发失败(' + r.status + ')'));
                    }
                    fillScPreview(data);
                    if (scPreviewPanel) {
                        scPreviewPanel.classList.remove('d-none');
                        scPreviewPanel.scrollIntoView({behavior: 'smooth', block: 'start'});
                    }
                })
                .catch((e) => showMsg('异常：' + e.message, false));
        });
    }

    if (formScAddConfirm) {
        formScAddConfirm.addEventListener('submit', function (e) {
            e.preventDefault();
            if (!scPreviewData) {
                setScPreviewMsg('请先执行 Add Preview', false);
                return;
            }

            let payload;
            try {
                payload = buildPayload();
            } catch (err) {
                setScPreviewMsg(err.message || '提交参数有误', false);
                return;
            }

            post('/api/triggers/sc/add', payload)
                .then(async (r) => {
                    const data = await r.json().catch(() => ({}));
                    if (!r.ok) {
                        throw new Error(data.error || ('触发失败(' + r.status + ')'));
                    }
                    setScPreviewMsg('AddSc 任务已启动，正在跳转任务列表', true);
                    showMsg('AddSc 任务已启动');
                    if (data && data.job_id) {
                        window.setTimeout(function () {
                            window.location.href = '/crawler/tasks?job_id=' + encodeURIComponent(data.job_id);
                        }, 300);
                    }
                })
                .catch((e) => setScPreviewMsg('异常：' + e.message, false));
        });
    }
})();
