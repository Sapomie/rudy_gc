(function () {
    const msgEl = document.getElementById('dSeedMsg');
    const openCreateBtn = document.getElementById('btnOpenDSeedCreateModal');
    const modalEl = document.getElementById('dSeedEditModal');
    const form = document.getElementById('dSeedEditForm');
    if (!window.bootstrap || !modalEl || !form) {
        return;
    }

    const modal = bootstrap.Modal.getOrCreateInstance(modalEl);
    const idEl = document.getElementById('dSeedEditId');
    const titleEl = document.getElementById('dSeedEditModalTitle');
    const metaEl = document.getElementById('dSeedEditModalMeta');
    const formMsgEl = document.getElementById('dSeedEditMsg');
    const submitBtn = document.getElementById('dSeedEditSubmit');
    const nameEl = document.getElementById('dSeedEditName');
    const activeEl = document.getElementById('dSeedEditActive');
    const searchTypeEl = document.getElementById('dSeedEditSearchType');
    const nameTypeEl = document.getElementById('dSeedEditNameType');
    const pageNowEl = document.getElementById('dSeedEditPageNow');
    const offsetEl = document.getElementById('dSeedEditOffset');
    const startPageEl = document.getElementById('dSeedEditStartPage');
    const endPageEl = document.getElementById('dSeedEditEndPage');
    const offsetWrapEl = document.getElementById('dSeedOffsetFieldWrap');
    const startPageWrapEl = document.getElementById('dSeedStartPageFieldWrap');
    const endPageWrapEl = document.getElementById('dSeedEndPageFieldWrap');
    const lastStatusEl = document.getElementById('dSeedEditLastStatus');
    const lastErrorEl = document.getElementById('dSeedEditLastError');

    function showMsg(el, text, ok) {
        if (!el) return;
        el.textContent = String(text || '');
        el.className = 'alert small py-2 px-3 mb-3 ' + (ok ? 'alert-success' : 'alert-danger');
    }

    function hideMsg(el) {
        if (!el) return;
        el.textContent = '';
        el.className = 'alert d-none mt-3 mb-0';
    }

    function setSubmitting(submitting) {
        if (!submitBtn) return;
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

    function setValue(el, value) {
        if (el) {
            el.value = String(value == null ? '' : value);
        }
    }

    function setHidden(el, hidden) {
        if (!el) return;
        el.classList.toggle('d-none', !!hidden);
    }

    function updateSearchTypeFields() {
        const searchType = String(searchTypeEl && searchTypeEl.value ? searchTypeEl.value : '1');
        const isOffset = searchType === '1';

        setHidden(offsetWrapEl, !isOffset);
        setHidden(startPageWrapEl, isOffset);
        setHidden(endPageWrapEl, isOffset);
    }

    function openCreateModal() {
        setValue(idEl, '');
        setValue(nameEl, '');
        setValue(activeEl, '2');
        setValue(searchTypeEl, '1');
        setValue(nameTypeEl, '1');
        setValue(pageNowEl, '0');
        setValue(offsetEl, '0');
        setValue(startPageEl, '0');
        setValue(endPageEl, '0');
        setValue(lastStatusEl, '0');
        setValue(lastErrorEl, '');
        if (titleEl) titleEl.textContent = '新增 d_seed 条目';
        if (metaEl) metaEl.textContent = '创建新的查询配置';
        hideMsg(formMsgEl);
        updateSearchTypeFields();
        setSubmitting(false);
        modal.show();
    }

    function openEditModal(button) {
        setValue(idEl, button.getAttribute('data-item-id'));
        setValue(nameEl, button.getAttribute('data-name'));
        setValue(activeEl, button.getAttribute('data-active'));
        setValue(searchTypeEl, button.getAttribute('data-search-type'));
        setValue(nameTypeEl, button.getAttribute('data-name-type'));
        setValue(pageNowEl, button.getAttribute('data-page-now'));
        setValue(offsetEl, button.getAttribute('data-offset'));
        setValue(startPageEl, button.getAttribute('data-start-page'));
        setValue(endPageEl, button.getAttribute('data-end-page'));
        setValue(lastStatusEl, button.getAttribute('data-last-status'));
        setValue(lastErrorEl, button.getAttribute('data-last-error'));
        if (titleEl) titleEl.textContent = '编辑 d_seed 条目';
        if (metaEl) metaEl.textContent = String(button.getAttribute('data-name') || '');
        hideMsg(formMsgEl);
        updateSearchTypeFields();
        setSubmitting(false);
        modal.show();
    }

    function buildPayload() {
        const searchType = Number(searchTypeEl.value || 0);
        const isOffset = searchType === 1;
        return {
            name: String(nameEl.value || '').trim(),
            active: Number(activeEl.value || 0),
            search_type: searchType,
            name_type: Number(nameTypeEl.value || 0),
            page_now: Number(pageNowEl.value || 0),
            offset: isOffset ? Number(offsetEl.value || 0) : 0,
            start_page: isOffset ? 0 : Number(startPageEl.value || 0),
            end_page: isOffset ? 0 : Number(endPageEl.value || 0),
            last_status: Number(lastStatusEl.value || 0),
            last_error: String(lastErrorEl.value || '').trim()
        };
    }

    if (openCreateBtn) {
        openCreateBtn.addEventListener('click', openCreateModal);
    }

    if (searchTypeEl) {
        searchTypeEl.addEventListener('change', updateSearchTypeFields);
    }

    document.addEventListener('click', function (event) {
        const editBtn = event.target.closest('.js-d-seed-edit');
        if (!editBtn) {
            return;
        }
        openEditModal(editBtn);
    });

    form.addEventListener('submit', function (event) {
        event.preventDefault();
        const itemID = String(idEl.value || '').trim();
        const payload = buildPayload();
        const isEdit = itemID !== '';
        const url = isEdit ? '/api/d-seeds/' + encodeURIComponent(itemID) : '/api/d-seeds';

        setSubmitting(true);
        hideMsg(formMsgEl);

        request(url, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(payload)
        }).then(function (data) {
            showMsg(msgEl, data.message || '保存成功', true);
            modal.hide();
            window.setTimeout(function () {
                window.location.reload();
            }, 250);
        }).catch(function (error) {
            showMsg(formMsgEl, error && error.message ? error.message : '保存失败', false);
            setSubmitting(false);
        });
    });

    updateSearchTypeFields();
}());
