(function () {
    const page = document.getElementById('wAggEventListPage');
    if (!page) {
        return;
    }

    const pageSize = document.getElementById('wAggEventPageSize');
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
