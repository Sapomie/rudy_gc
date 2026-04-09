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

    const btnScAddPreview = document.getElementById('btnScAddPreview');
    const modalEl = document.getElementById('addScPreviewModal');
    const formScAddConfirm = document.getElementById('formScAddConfirm');
    const scPreviewMsg = document.getElementById('scAddPreviewMsg');
    const scPreviewName = document.getElementById('scPreviewName');
    const scPreviewImage = document.getElementById('scPreviewImage');
    const scPreviewMovieCount = document.getElementById('scPreviewMovieCount');
    const scPreviewComeMovie = document.getElementById('scPreviewComeMovie');
    const scPreviewMovieCast = document.getElementById('scPreviewMovieCast');
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

    function buildCastOptions(movieJavId) {
        if (!scPreviewMovieCast) return;
        scPreviewMovieCast.innerHTML = '';
        if (!scPreviewData || !Array.isArray(scPreviewData.movies)) return;
        const movie = scPreviewData.movies.find((item) => item.movieJavId === movieJavId);
        if (!movie || !Array.isArray(movie.casts) || movie.casts.length === 0) {
            const option = document.createElement('option');
            option.value = '';
            option.textContent = '无演员可选';
            scPreviewMovieCast.appendChild(option);
            return;
        }
        movie.casts.forEach((name, idx) => {
            const option = document.createElement('option');
            option.value = name;
            option.textContent = name;
            if (idx === 0) option.selected = true;
            scPreviewMovieCast.appendChild(option);
        });
    }

    function fillScPreview(data) {
        scPreviewData = data;
        if (scPreviewName) scPreviewName.value = data.scName || '';
        if (scPreviewImage) scPreviewImage.value = data.imageFound ? (data.imageName || '已发现图片') : '无图片';
        if (scPreviewMovieCount) scPreviewMovieCount.value = String(data.movieCount || 0);
        if (scPreviewDuration) scPreviewDuration.value = '';
        if (scPreviewFg) scPreviewFg.value = '';
        if (scPreviewVessel) scPreviewVessel.value = '';
        if (scPreviewRemarks) scPreviewRemarks.value = '';
        setScPreviewMsg('', true);

        if (scPreviewComeMovie) {
            scPreviewComeMovie.innerHTML = '';
            (data.movies || []).forEach((movie, idx) => {
                const option = document.createElement('option');
                option.value = movie.movieJavId;
                option.textContent = movie.movieName;
                if (idx === 0) option.selected = true;
                scPreviewComeMovie.appendChild(option);
            });
            buildCastOptions(scPreviewComeMovie.value);
        }
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
                    if (modalEl && window.bootstrap && window.bootstrap.Modal) {
                        window.bootstrap.Modal.getOrCreateInstance(modalEl).show();
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
	            const payload = {
	                comeMovieJavId: (scPreviewComeMovie && scPreviewComeMovie.value || '').trim(),
	                movieCast: (scPreviewMovieCast && scPreviewMovieCast.value || '').trim(),
	                fg: (scPreviewFg && scPreviewFg.value || '').trim(),
	                vessel: (scPreviewVessel && scPreviewVessel.value || '').trim(),
	                remarks: (scPreviewRemarks && scPreviewRemarks.value || '').trim(),
	            };
	            const rawDurationMinutes = (scPreviewDuration && scPreviewDuration.value || '').trim();
	            if (rawDurationMinutes === '') {
	                payload.duration = 0;
	            } else {
	                const parsedDurationMinutes = parseInt(rawDurationMinutes, 10);
	                if (!Number.isFinite(parsedDurationMinutes)) {
	                    setScPreviewMsg('时长(分钟)必须是数字', false);
	                    return;
	                }
	                payload.duration = parsedDurationMinutes;
	            }
	            if (!payload.comeMovieJavId) {
	                setScPreviewMsg('请选择 Come Movie', false);
	                return;
            }
            if (!payload.movieCast) {
                setScPreviewMsg('请选择 Movie Cast', false);
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
	                    if (modalEl && window.bootstrap && window.bootstrap.Modal) {
	                        window.bootstrap.Modal.getOrCreateInstance(modalEl).hide();
	                    }
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
