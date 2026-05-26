(function () {
    const pageMsgEl = document.getElementById('scEventPageMsg');
    const modalEl = document.getElementById('scEventEditModal');
    const formEl = document.getElementById('scEventEditForm');
    const submitBtn = document.getElementById('scEventEditSubmit');
    const formMsgEl = document.getElementById('scEventEditMsg');
    const comeMovieEl = document.getElementById('scEventEditComeMovie');
    const castSourcePoolEl = document.getElementById('scEventEditCastSourcePool');
    const kindEl = document.getElementById('scEventEditKind');
    const durationEl = document.getElementById('scEventEditDuration');
    const fgEl = document.getElementById('scEventEditFg');
    const vesselEl = document.getElementById('scEventEditVessel');
    const movieCastEl = document.getElementById('scEventEditMovieCast');
    const remarksEl = document.getElementById('scEventEditRemarks');

    if (!formEl || !submitBtn || !comeMovieEl || !castSourcePoolEl || !kindEl || !durationEl || !fgEl || !vesselEl || !movieCastEl || !remarksEl) {
        return;
    }

    function showMsg(el, text, ok, extraClass) {
        if (!el) return;
        el.textContent = String(text || '');
        el.className = 'alert ' + (extraClass ? extraClass + ' ' : '') + (ok ? 'alert-success' : 'alert-danger');
    }

    function hideMsg(el, extraClass) {
        if (!el) return;
        el.textContent = '';
        el.className = 'alert d-none' + (extraClass ? (' ' + extraClass) : '');
    }

    function setSubmitting(submitting) {
        submitBtn.disabled = submitting;
        submitBtn.textContent = submitting ? '保存中...' : '保存';
    }

    function request(url, options) {
        return fetch(url, options).then(async function (response) {
            const data = await response.json().catch(function () { return {}; });
            if (!response.ok || !data || data.ok === false) {
                throw new Error((data && data.error) || ('请求失败(' + response.status + ')'));
            }
            return data;
        });
    }

    function listCastNames(movieJavId) {
        const normalizedMovieJavId = String(movieJavId || '').trim();
        if (!normalizedMovieJavId) {
            return [];
        }
        const sourceEls = castSourcePoolEl.querySelectorAll('.js-sc-event-cast-source');
        for (let i = 0; i < sourceEls.length; i += 1) {
            const sourceEl = sourceEls[i];
            if (!sourceEl || String(sourceEl.getAttribute('data-movie-jav-id') || '').trim() !== normalizedMovieJavId) {
                continue;
            }
            return Array.from(sourceEl.querySelectorAll('.js-sc-event-cast-item'))
                .map((itemEl) => String(itemEl.getAttribute('data-cast-name') || '').trim())
                .filter((name) => name !== '');
        }
        return [];
    }

    function appendMovieCastOption(value, selected) {
        const option = document.createElement('option');
        option.value = value;
        option.textContent = value;
        option.selected = !!selected;
        movieCastEl.appendChild(option);
    }

    function rebuildMovieCastOptions(movieJavId, preferredValue) {
        const normalizedPreferredValue = String(preferredValue || '').trim();
        const castNames = listCastNames(movieJavId);
        movieCastEl.innerHTML = '';

        const emptyOption = document.createElement('option');
        emptyOption.value = '';
        emptyOption.textContent = '可为空';
        emptyOption.selected = normalizedPreferredValue === '';
        movieCastEl.appendChild(emptyOption);

        let matched = normalizedPreferredValue === '';
        castNames.forEach(function (name) {
            const selected = normalizedPreferredValue !== '' && name === normalizedPreferredValue;
            if (selected) {
                matched = true;
            }
            appendMovieCastOption(name, selected);
        });

        if (!matched && normalizedPreferredValue !== '') {
            appendMovieCastOption(normalizedPreferredValue, true);
        }
    }

    function buildPayload() {
        return {
            comeMovieJavId: String(comeMovieEl.value || '').trim(),
            kind: String(kindEl.value || '').trim(),
            duration: Number(durationEl.value || 0),
            fg: String(fgEl.value || '').trim(),
            vessel: String(vesselEl.value || '').trim(),
            movieCast: String(movieCastEl.value || '').trim(),
            remarks: String(remarksEl.value || '').trim()
        };
    }

    comeMovieEl.addEventListener('change', function () {
        rebuildMovieCastOptions(comeMovieEl.value, '');
    });

    formEl.addEventListener('submit', function (event) {
        event.preventDefault();
        hideMsg(formMsgEl, 'mt-3 mb-0');
        setSubmitting(true);

        request(String(formEl.getAttribute('data-update-url') || ''), {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(buildPayload())
        }).then(function (data) {
            showMsg(pageMsgEl, data.message || '保存成功', true, 'mb-3');
            if (window.bootstrap && modalEl) {
                const modal = bootstrap.Modal.getOrCreateInstance(modalEl);
                modal.hide();
            }
            window.setTimeout(function () {
                window.location.reload();
            }, 250);
        }).catch(function (error) {
            showMsg(formMsgEl, error && error.message ? error.message : '保存失败', false, 'mt-3 mb-0');
            setSubmitting(false);
        });
    });

    rebuildMovieCastOptions(comeMovieEl.value, movieCastEl.getAttribute('data-selected-value') || movieCastEl.value || '');
}());
