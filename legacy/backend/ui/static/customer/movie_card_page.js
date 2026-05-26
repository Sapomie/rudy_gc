(function () {
    function initMovieCardPageDateEditor() {
        const openButton = document.querySelector('.js-open-page-date-modal');
        const modalEl = document.getElementById('movieCardPageDateModal');
        const form = document.getElementById('movieCardPageDateForm');
        const titleEl = document.getElementById('movieCardPageDateTitle');
        const keyEl = document.getElementById('movieCardPageDateKey');
        const inputEl = document.getElementById('movieCardPageDateInput');
        const todayBtn = document.getElementById('movieCardPageDateTodayBtn');
        const msgEl = document.getElementById('movieCardPageDateMsg');
        const submitBtn = document.getElementById('movieCardPageDateSubmit');

        if (!openButton || !modalEl || !form || !window.bootstrap || !window.bootstrap.Modal) {
            return;
        }

        const modal = window.bootstrap.Modal.getOrCreateInstance(modalEl);

        function showMsg(text, ok) {
            if (!msgEl) {
                return;
            }
            msgEl.textContent = String(text || '');
            msgEl.className = 'alert mb-0 ' + (ok ? 'alert-success' : 'alert-danger');
        }

        function hideMsg() {
            if (!msgEl) {
                return;
            }
            msgEl.textContent = '';
            msgEl.className = 'alert d-none mb-0';
        }

        function setSubmitting(submitting) {
            if (!submitBtn) {
                return;
            }
            submitBtn.disabled = submitting;
            submitBtn.textContent = submitting ? '保存中...' : '保存';
        }

        function normalizeDateInputValue(raw) {
            const digits = String(raw || '').replace(/\D/g, '').slice(0, 8);
            if (digits.length <= 4) {
                return digits;
            }
            if (digits.length <= 6) {
                return digits.slice(0, 4) + '-' + digits.slice(4);
            }
            return digits.slice(0, 4) + '-' + digits.slice(4, 6) + '-' + digits.slice(6);
        }

        function isValidDateValue(raw) {
            return /^[0-9]{4}-[0-9]{2}-[0-9]{2}$/.test(String(raw || '').trim());
        }

        function todayDateString() {
            const now = new Date();
            const year = String(now.getFullYear());
            const month = String(now.getMonth() + 1).padStart(2, '0');
            const day = String(now.getDate()).padStart(2, '0');
            return year + '-' + month + '-' + day;
        }

        if (inputEl) {
            inputEl.addEventListener('input', function () {
                const normalized = normalizeDateInputValue(inputEl.value);
                if (inputEl.value !== normalized) {
                    inputEl.value = normalized;
                }
            });
        }

        if (todayBtn && inputEl) {
            todayBtn.addEventListener('click', function () {
                inputEl.value = todayDateString();
                hideMsg();
                inputEl.focus();
            });
        }

        openButton.addEventListener('click', function () {
            const itemKey = String(openButton.getAttribute('data-item-key') || '').trim();
            const itemLabel = String(openButton.getAttribute('data-item-label') || '').trim() || '日期';
            const itemValue = String(openButton.getAttribute('data-item-value') || '').trim();
            if (!itemKey || !inputEl || !keyEl) {
                return;
            }
            if (titleEl) {
                titleEl.textContent = '修改' + itemLabel;
            }
            keyEl.value = itemKey;
            inputEl.value = itemValue;
            hideMsg();
            modal.show();
        });

        form.addEventListener('submit', function (event) {
            event.preventDefault();
            const itemKey = String(keyEl && keyEl.value || '').trim();
            const itemValue = String(inputEl && inputEl.value || '').trim();
            if (!itemKey) {
                showMsg('缺少日期 key', false);
                return;
            }
            if (!itemValue) {
                showMsg('请选择日期', false);
                return;
            }
            if (!isValidDateValue(itemValue)) {
                showMsg('日期格式必须为 YYYY-MM-DD', false);
                return;
            }

            setSubmitting(true);
            fetch('/api/w-kv/date', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    item_key: itemKey,
                    item_value: itemValue
                })
            }).then(async function (response) {
                const data = await response.json().catch(function () {
                    return {};
                });
                if (!response.ok) {
                    throw new Error(data.error || ('请求失败(' + response.status + ')'));
                }
                return data;
            }).then(function (data) {
                const savedValue = data && data.item_value ? String(data.item_value).trim() : itemValue;
                const itemLabel = String(openButton.getAttribute('data-item-label') || '').trim() || '日期';
                openButton.setAttribute('data-item-value', savedValue);
                openButton.textContent = itemLabel + '：' + savedValue;
                hideMsg();
                setSubmitting(false);
                modal.hide();
            }).catch(function (error) {
                showMsg(error && error.message ? error.message : '保存失败', false);
                setSubmitting(false);
            });
        });
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initMovieCardPageDateEditor);
        return;
    }
    initMovieCardPageDateEditor();
}());
