(function () {
    const page = document.getElementById('wMediaBirthBucketListPage');
    if (!page) {
        return;
    }

    const pageSize = document.getElementById('bucketPageSize');
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
