(function () {
    const API = {
        aggBackfill: '/api/agg/w-media/backfill',
        releaseAggBackfill: '/api/agg/movie-release/backfill',
    };

    const btnBackfillWMediaAgg = document.getElementById('btnBackfillWMediaAgg');
    const btnBackfillMovieReleaseAgg = document.getElementById('btnBackfillMovieReleaseAgg');
    const aggBackfillStatusText = document.getElementById('aggBackfillStatusText');
    const releaseAggBackfillStatusText = document.getElementById('releaseAggBackfillStatusText');
    const msgArea = document.getElementById('msgArea');

    const aggClearedBucketRows = document.getElementById('aggClearedBucketRows');
    const aggClearedTopRows = document.getElementById('aggClearedTopRows');
    const aggClearedDirtyRows = document.getElementById('aggClearedDirtyRows');
    const aggDirtyDays = document.getElementById('aggDirtyDays');
    const aggYearBuckets = document.getElementById('aggYearBuckets');
    const aggQuarterBuckets = document.getElementById('aggQuarterBuckets');
    const aggMonthBuckets = document.getElementById('aggMonthBuckets');
    const aggDayBuckets = document.getElementById('aggDayBuckets');
    const aggTopRows = document.getElementById('aggTopRows');
    const aggElapsedMs = document.getElementById('aggElapsedMs');

    const releaseAggClearedBucketRows = document.getElementById('releaseAggClearedBucketRows');
    const releaseAggClearedTopRows = document.getElementById('releaseAggClearedTopRows');
    const releaseAggClearedDirtyRows = document.getElementById('releaseAggClearedDirtyRows');
    const releaseAggDirtyMonths = document.getElementById('releaseAggDirtyMonths');
    const releaseAggYearBuckets = document.getElementById('releaseAggYearBuckets');
    const releaseAggQuarterBuckets = document.getElementById('releaseAggQuarterBuckets');
    const releaseAggMonthBuckets = document.getElementById('releaseAggMonthBuckets');
    const releaseAggDayBuckets = document.getElementById('releaseAggDayBuckets');
    const releaseAggTopRows = document.getElementById('releaseAggTopRows');
    const releaseAggElapsedMs = document.getElementById('releaseAggElapsedMs');

    let loading = false;
    let msgTimer = null;

    function showMsg(text, ok) {
        if (!msgArea) {
            return;
        }
        clearTimeout(msgTimer);
        msgArea.textContent = text;
        msgArea.className = 'alert ' + (ok ? 'alert-success' : 'alert-danger') + ' fade show small py-2 px-3';
        msgArea.style.display = 'block';
        msgTimer = setTimeout(function () {
            msgArea.style.display = 'none';
        }, 3000);
    }

    function setLoading(next) {
        loading = !!next;
        if (btnBackfillWMediaAgg) btnBackfillWMediaAgg.disabled = loading;
        if (btnBackfillMovieReleaseAgg) btnBackfillMovieReleaseAgg.disabled = loading;
    }

    function requestAggBackfill() {
        return fetch(API.aggBackfill, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: '{}',
        }).then(async function (resp) {
            const payload = await resp.json().catch(function () {
                return {};
            });
            const result = payload && payload.result ? payload.result : null;
            if (!resp.ok) {
                const err = new Error((payload && payload.error) || ('请求失败(' + resp.status + ')'));
                err.result = result && result.result ? result.result : result;
                throw err;
            }
            return result && result.result ? result.result : result;
        });
    }

    function requestReleaseAggBackfill() {
        return fetch(API.releaseAggBackfill, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: '{}',
        }).then(async function (resp) {
            const payload = await resp.json().catch(function () {
                return {};
            });
            const result = payload && payload.result ? payload.result : null;
            if (!resp.ok) {
                const err = new Error((payload && payload.error) || ('请求失败(' + resp.status + ')'));
                err.result = result && result.result ? result.result : result;
                throw err;
            }
            return result && result.result ? result.result : result;
        });
    }

    function renderAggBackfillState(data) {
        const safe = data || {};
        if (aggClearedBucketRows) aggClearedBucketRows.textContent = String(safe.cleared_bucket_rows || 0);
        if (aggClearedTopRows) aggClearedTopRows.textContent = String(safe.cleared_top_rows || 0);
        if (aggClearedDirtyRows) aggClearedDirtyRows.textContent = String(safe.cleared_dirty_rows || 0);
        if (aggDirtyDays) aggDirtyDays.textContent = String(safe.dirty_days || 0);
        if (aggYearBuckets) aggYearBuckets.textContent = String(safe.year_buckets || 0);
        if (aggQuarterBuckets) aggQuarterBuckets.textContent = String(safe.quarter_buckets || 0);
        if (aggMonthBuckets) aggMonthBuckets.textContent = String(safe.month_buckets || 0);
        if (aggDayBuckets) aggDayBuckets.textContent = String(safe.day_buckets || 0);
        if (aggTopRows) aggTopRows.textContent = String(safe.top_rows || 0);
        if (aggElapsedMs) aggElapsedMs.textContent = String(safe.elapsed_ms || 0) + ' ms';
        if (aggBackfillStatusText) {
            if (!safe || Object.keys(safe).length === 0) {
                aggBackfillStatusText.textContent = '尚未执行回填';
            } else {
                aggBackfillStatusText.textContent = '回填完成：日桶 ' + String(safe.day_buckets || 0) + '，Top 行 ' + String(safe.top_rows || 0);
            }
        }
    }

    function renderReleaseAggBackfillState(data) {
        const safe = data || {};
        if (releaseAggClearedBucketRows) releaseAggClearedBucketRows.textContent = String(safe.cleared_bucket_rows || 0);
        if (releaseAggClearedTopRows) releaseAggClearedTopRows.textContent = String(safe.cleared_top_rows || 0);
        if (releaseAggClearedDirtyRows) releaseAggClearedDirtyRows.textContent = String(safe.cleared_dirty_rows || 0);
        if (releaseAggDirtyMonths) releaseAggDirtyMonths.textContent = String(safe.dirty_months || 0);
        if (releaseAggYearBuckets) releaseAggYearBuckets.textContent = String(safe.year_buckets || 0);
        if (releaseAggQuarterBuckets) releaseAggQuarterBuckets.textContent = String(safe.quarter_buckets || 0);
        if (releaseAggMonthBuckets) releaseAggMonthBuckets.textContent = String(safe.month_buckets || 0);
        if (releaseAggDayBuckets) releaseAggDayBuckets.textContent = String(safe.day_buckets || 0);
        if (releaseAggTopRows) releaseAggTopRows.textContent = String(safe.top_rows || 0);
        if (releaseAggElapsedMs) releaseAggElapsedMs.textContent = String(safe.elapsed_ms || 0) + ' ms';
        if (releaseAggBackfillStatusText) {
            if (!safe || Object.keys(safe).length === 0) {
                releaseAggBackfillStatusText.textContent = '尚未执行回填';
            } else {
                releaseAggBackfillStatusText.textContent = '回填完成：日桶 ' + String(safe.day_buckets || 0) + '，Top 行 ' + String(safe.top_rows || 0);
            }
        }
    }

    if (btnBackfillWMediaAgg) {
        btnBackfillWMediaAgg.addEventListener('click', function () {
            setLoading(true);
            if (aggBackfillStatusText) {
                aggBackfillStatusText.textContent = '正在回填旧的 WMedia 时间聚合数据...';
            }
            requestAggBackfill()
                .then(function (data) {
                    renderAggBackfillState(data);
                    showMsg('WMedia 时间聚合回填完成', true);
                })
                .catch(function (err) {
                    showMsg(err.message || 'WMedia 时间聚合回填失败', false);
                    if (err.result) {
                        renderAggBackfillState(err.result);
                    }
                    if (aggBackfillStatusText) {
                        aggBackfillStatusText.textContent = '回填失败，请查看提示信息';
                    }
                })
                .finally(function () {
                    setLoading(false);
                });
        });
    }

    if (btnBackfillMovieReleaseAgg) {
        btnBackfillMovieReleaseAgg.addEventListener('click', function () {
            setLoading(true);
            if (releaseAggBackfillStatusText) {
                releaseAggBackfillStatusText.textContent = '正在回填...';
            }
            requestReleaseAggBackfill()
                .then(function (data) {
                    renderReleaseAggBackfillState(data);
                    showMsg('上映日时间聚合回填完成', true);
                })
                .catch(function (err) {
                    showMsg(err.message || '上映日时间聚合回填失败', false);
                    if (err.result) {
                        renderReleaseAggBackfillState(err.result);
                    }
                })
                .finally(function () {
                    setLoading(false);
                });
        });
    }

    renderAggBackfillState(null);
    renderReleaseAggBackfillState(null);
})();
