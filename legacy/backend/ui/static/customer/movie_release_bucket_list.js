(function () {
    const page = document.getElementById('movieReleaseBucketListPage');
    if (!page) {
        return;
    }

    const pageSize = document.getElementById('releaseBucketPageSize');
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
