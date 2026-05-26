(function () {
    const page = document.getElementById('crawlRecordListPage');
    if (!page) {
        return;
    }

    const pageSize = document.getElementById('crawlRecordPageSize');
    if (!pageSize) {
        return;
    }

    pageSize.addEventListener('change', function () {
        const form = pageSize.closest('form');
        if (form) {
            form.submit();
        }
    });
})();
