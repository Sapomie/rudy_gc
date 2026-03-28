(function () {
    const msgArea = document.getElementById('castMsgArea');
    const btn = document.getElementById('btnRebuildActorRank');
    const main = document.querySelector('.main-content');
    const personId = parseInt((main && main.dataset.personId) || '0', 10) || 0;
    const mergeSearchInput = document.getElementById('mergeSearchInput');
    const btnMergeSearch = document.getElementById('btnMergeSearch');
    const mergeCandidateList = document.getElementById('mergeCandidateList');
    const mergeSelectedList = document.getElementById('mergeSelectedList');
    const btnPreviewMerge = document.getElementById('btnPreviewMerge');
    const mergePreviewArea = document.getElementById('mergePreviewArea');
    const mergeConfirmModal = document.getElementById('personMergeConfirmModal');
    const mergeConfirmSummary = document.getElementById('mergeConfirmSummary');
    const btnExecuteMerge = document.getElementById('btnExecuteMerge');
    const btnToggleScHistory = document.getElementById('btnToggleScHistory');
    const scHistoryExtraRows = Array.from(document.querySelectorAll('.sc-history-extra-row'));
    const selectedSources = new Map();
    let currentPreview = null;
    let scHistoryExpanded = false;

    function showMsg(text, ok = true) {
        if (!msgArea) return;
        msgArea.textContent = text;
        msgArea.className = 'alert ' + (ok ? 'alert-success' : 'alert-danger') + ' small py-2 px-3 mb-3';
        msgArea.style.display = 'block';
    }

    function escapeHtml(text) {
        return String(text || '')
            .replaceAll('&', '&amp;')
            .replaceAll('<', '&lt;')
            .replaceAll('>', '&gt;')
            .replaceAll('"', '&quot;')
            .replaceAll("'", '&#39;');
    }

    function normalizePerson(person) {
        const row = person || {};
        return {
            id: row.id || row.Id || 0,
            name: row.name || row.Name || '',
            chinese: row.chinese || row.Chinese || '',
            birthDay: row.birthDay || row.BirthDay || 0,
            movieNumber: row.movieNumber || row.MovieNumber || 0,
            ownedMovieNumber: row.ownedMovieNumber || row.OwnedMovieNumber || 0,
            scTimes: row.scTimes || row.ScTimes || 0,
            comeTimes: row.comeTimes || row.ComeTimes || 0,
        };
    }

    function personDisplayName(person) {
        const row = normalizePerson(person);
        return row.chinese || row.name || '(无名 person)';
    }

    function formatLocalDate(sec) {
        if (!Number.isFinite(sec) || sec <= 0) {
            return '';
        }
        const date = new Date(sec * 1000);
        const yyyy = date.getFullYear();
        const mm = String(date.getMonth() + 1).padStart(2, '0');
        const dd = String(date.getDate()).padStart(2, '0');
        return yyyy + '-' + mm + '-' + dd;
    }

    function renderNameBadges(names) {
        if (!Array.isArray(names) || names.length === 0) {
            return '<span class="text-muted small">无 cast</span>';
        }
        return names.map((name) => '<span class="cast-alias-badge">' + escapeHtml(name) + '</span>').join('');
    }

    function renderCandidateRow(candidate, selected) {
        const person = normalizePerson(candidate && candidate.person ? candidate.person : candidate && candidate.Person);
        const castNames = Array.isArray(candidate && candidate.castNames) ? candidate.castNames : [];
        const primaryName = personDisplayName(person);
        const rawName = person.chinese && person.name ? ('原名：' + escapeHtml(person.name)) : '';
        const birthDay = person.birthDay > 0 ? ('生日：' + formatLocalDate(person.birthDay)) : '';
        const metrics = [
            'Movies ' + (person.movieNumber || 0),
            'Owned ' + (person.ownedMovieNumber || 0),
            'SC ' + (person.scTimes || 0),
            'COME ' + (person.comeTimes || 0),
        ].join(' / ');

        return '' +
            '<div class="merge-person-item' + (selected ? ' is-selected' : '') + '" data-person-id="' + escapeHtml(person.id) + '" data-person-name="' + escapeHtml(person.name) + '" data-person-chinese="' + escapeHtml(person.chinese) + '">' +
            '<div class="merge-person-head">' +
            '<div>' +
            '<div class="merge-person-name">' + escapeHtml(primaryName) + ' <span class="text-muted small">#' + escapeHtml(person.id) + '</span></div>' +
            '<div class="merge-person-meta">' +
            (rawName ? '<span>' + rawName + '</span>' : '') +
            (birthDay ? '<span>' + birthDay + '</span>' : '') +
            '<span>' + escapeHtml(metrics) + '</span>' +
            '</div>' +
            '<div class="cast-alias-wrap">' + renderNameBadges(castNames) + '</div>' +
            '</div>' +
            '<button type="button" class="btn btn-sm ' + (selected ? 'btn-outline-secondary js-remove-source' : 'btn-outline-primary js-add-source') + '" data-person-id="' + escapeHtml(person.id) + '">' +
            (selected ? '移除' : '加入') +
            '</button>' +
            '</div>' +
            '</div>';
    }

    function renderCandidateList(candidates) {
        if (!mergeCandidateList) return;
        if (!Array.isArray(candidates) || candidates.length === 0) {
            mergeCandidateList.innerHTML = '<div class="merge-list-empty">没有找到可合并的候选 person。</div>';
            return;
        }
        mergeCandidateList.innerHTML = candidates
            .map((candidate) => {
                const person = normalizePerson(candidate && candidate.person ? candidate.person : candidate && candidate.Person);
                return renderCandidateRow(candidate, selectedSources.has(person.id));
            })
            .join('');
    }

    function renderSelectedList() {
        if (!mergeSelectedList) return;
        const values = Array.from(selectedSources.values());
        if (values.length === 0) {
            mergeSelectedList.innerHTML = '<div class="merge-list-empty">还没有选择来源 person。</div>';
            if (btnPreviewMerge) btnPreviewMerge.disabled = true;
            return;
        }
        mergeSelectedList.innerHTML = values.map((candidate) => renderCandidateRow(candidate, true)).join('');
        if (btnPreviewMerge) btnPreviewMerge.disabled = false;
    }

    function collectSourcePersonIds() {
        return Array.from(selectedSources.keys()).map((id) => Number(id)).filter((id) => Number.isFinite(id) && id > 0);
    }

    function request(url, options) {
        return fetch(url, options).then(async (resp) => {
            const data = await resp.json().catch(() => ({}));
            if (!resp.ok) {
                throw new Error(data.error || ('请求失败(' + resp.status + ')'));
            }
            return data;
        });
    }

    function renderPreview(preview) {
        if (!mergePreviewArea) return;
        currentPreview = preview || null;
        if (!preview || !preview.keep) {
            mergePreviewArea.style.display = 'none';
            mergePreviewArea.innerHTML = '';
            return;
        }

        const sourceHtml = (preview.sources || []).map((item) => {
            const person = normalizePerson(item.person || item.Person);
            return '' +
                '<div class="merge-preview-box">' +
                '<div class="merge-preview-label">来源 Person #' + escapeHtml(person.id) + '</div>' +
                '<div class="merge-person-name">' + escapeHtml(personDisplayName(person)) + '</div>' +
                '<div class="cast-alias-wrap mt-2">' + renderNameBadges(item.castNames || []) + '</div>' +
                '</div>';
        }).join('');

        const keepPerson = normalizePerson(preview.keep.person || preview.keep.Person);
        mergePreviewArea.innerHTML = '' +
            '<div class="merge-preview-box">' +
            '<div class="merge-preview-label">保留 Person #' + escapeHtml(keepPerson.id) + '</div>' +
            '<div class="merge-person-name">' + escapeHtml(personDisplayName(keepPerson)) + '</div>' +
            '<div class="cast-alias-wrap mt-2">' + renderNameBadges(preview.keep.castNames || []) + '</div>' +
            '</div>' +
            (sourceHtml || '') +
            '<div class="merge-preview-box">' +
            '<div class="merge-preview-label">本次迁移</div>' +
            '<div class="cast-alias-wrap mb-2">' + renderNameBadges(preview.moveCastNames || []) + '</div>' +
            '<div class="small text-muted">删除 person：' + escapeHtml((preview.removePersonIds || []).join(', ') || '-') + '</div>' +
            '<div class="small text-muted mt-1">受影响影片：' + escapeHtml(preview.affectedMovieCount || 0) + '</div>' +
            '<div class="merge-preview-actions">' +
            '<button type="button" class="btn btn-danger" id="btnOpenMergeConfirm">执行合并</button>' +
            '</div>' +
            '</div>';
        mergePreviewArea.style.display = 'grid';

        const btnOpenMergeConfirm = document.getElementById('btnOpenMergeConfirm');
        if (btnOpenMergeConfirm) {
            btnOpenMergeConfirm.addEventListener('click', function () {
                if (!mergeConfirmSummary || !mergeConfirmModal || !window.bootstrap || !window.bootstrap.Modal) return;
                mergeConfirmSummary.innerHTML = '' +
                    '<div>Keep Person #' + escapeHtml(keepPerson.id) + '</div>' +
                    '<div class="mt-1">删除 Person：' + escapeHtml((preview.removePersonIds || []).join(', ') || '-') + '</div>' +
                    '<div class="mt-1">迁移 Cast：' + escapeHtml((preview.moveCastNames || []).join(', ') || '-') + '</div>' +
                    '<div class="mt-1">受影响影片：' + escapeHtml(preview.affectedMovieCount || 0) + '</div>';
                window.bootstrap.Modal.getOrCreateInstance(mergeConfirmModal).show();
            });
        }
    }

    function searchCandidates() {
        if (personId <= 0 || !mergeSearchInput) return;
        const keyword = (mergeSearchInput.value || '').trim();
        if (!keyword) {
            renderCandidateList([]);
            showMsg('请输入搜索关键词', false);
            return;
        }

        request('/api/persons/' + encodeURIComponent(personId) + '/merge-candidates?q=' + encodeURIComponent(keyword), {method: 'GET'})
            .then((data) => {
                renderCandidateList(data.candidates || []);
            })
            .catch((err) => showMsg('搜索失败：' + err.message, false));
    }

    if (btn) {
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
    }

    if (btnMergeSearch) {
        btnMergeSearch.addEventListener('click', searchCandidates);
    }

    if (mergeSearchInput) {
        mergeSearchInput.addEventListener('keydown', function (event) {
            if (event.key === 'Enter') {
                event.preventDefault();
                searchCandidates();
            }
        });
    }

    if (btnToggleScHistory && scHistoryExtraRows.length > 0) {
        btnToggleScHistory.addEventListener('click', function () {
            scHistoryExpanded = !scHistoryExpanded;
            scHistoryExtraRows.forEach(function (row) {
                row.classList.toggle('d-none', !scHistoryExpanded);
            });
            btnToggleScHistory.textContent = scHistoryExpanded
                ? (btnToggleScHistory.dataset.collapseText || '收起')
                : (btnToggleScHistory.dataset.expandText || '展开全部');
            btnToggleScHistory.setAttribute('aria-expanded', scHistoryExpanded ? 'true' : 'false');
        });
    }

    if (mergeCandidateList) {
        mergeCandidateList.addEventListener('click', function (event) {
            const button = event.target.closest('.js-add-source, .js-remove-source');
            if (!button) return;
            const personIdValue = parseInt(button.dataset.personId || '0', 10);
            if (!Number.isFinite(personIdValue) || personIdValue <= 0) return;

            const row = button.closest('.merge-person-item');
            if (!row) return;

            const castEls = row.querySelectorAll('.cast-alias-badge');
            const castNames = Array.from(castEls).map((el) => (el.textContent || '').trim()).filter(Boolean);
            const candidate = {
                person: {
                    id: personIdValue,
                    name: (row.dataset.personName || '').trim(),
                    chinese: (row.dataset.personChinese || '').trim(),
                },
                castNames,
            };

            if (selectedSources.has(personIdValue)) {
                selectedSources.delete(personIdValue);
            } else {
                selectedSources.set(personIdValue, candidate);
            }

            renderSelectedList();
            currentPreview = null;
            renderPreview(null);
            searchCandidates();
        });
    }

    if (mergeSelectedList) {
        mergeSelectedList.addEventListener('click', function (event) {
            const button = event.target.closest('.js-remove-source');
            if (!button) return;
            const personIdValue = parseInt(button.dataset.personId || '0', 10);
            if (!Number.isFinite(personIdValue) || personIdValue <= 0) return;
            selectedSources.delete(personIdValue);
            renderSelectedList();
            currentPreview = null;
            renderPreview(null);
            searchCandidates();
        });
    }

    if (btnPreviewMerge) {
        btnPreviewMerge.addEventListener('click', function () {
            const sourcePersonIds = collectSourcePersonIds();
            if (sourcePersonIds.length === 0) {
                showMsg('请先选择来源 person', false);
                return;
            }

            request('/api/persons/merge-preview', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    keepPersonId: personId,
                    sourcePersonIds,
                }),
            })
                .then((data) => {
                    renderPreview(data.preview);
                    showMsg('合并预览已生成');
                })
                .catch((err) => showMsg('预览失败：' + err.message, false));
        });
    }

    if (btnExecuteMerge) {
        btnExecuteMerge.addEventListener('click', function () {
            if (!currentPreview) {
                showMsg('请先生成预览', false);
                return;
            }

            request('/api/persons/merge', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    keepPersonId: personId,
                    sourcePersonIds: collectSourcePersonIds(),
                }),
            })
                .then((data) => {
                    if (mergeConfirmModal && window.bootstrap && window.bootstrap.Modal) {
                        window.bootstrap.Modal.getOrCreateInstance(mergeConfirmModal).hide();
                    }
                    showMsg('MergePerson 已完成，正在刷新页面');
                    window.setTimeout(function () {
                        window.location.href = '/cast?id=' + encodeURIComponent(data.result.keepPersonId || personId);
                    }, 300);
                })
                .catch((err) => showMsg('执行失败：' + err.message, false));
        });
    }

    renderSelectedList();
})();
