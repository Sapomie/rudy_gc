(function (global) {
    const page = document.getElementById('crawlerDetailLoopPage');
    if (!page) {
        return;
    }

    const API_BASE = (global.CRAWLER_DETAIL_LOOP_PAGE_DATA && global.CRAWLER_DETAIL_LOOP_PAGE_DATA.apiBase) || '/api/crawler/detail-loop';
    const EVENT_LIMIT = 500;

    const state = {
        events: [],
        eventCount: 0,
        eventSource: null,
        statusTimer: null,
        seenMap: Object.create(null),
    };

    const startBtn = document.getElementById('startBtn');
    const stopBtn = document.getElementById('stopBtn');
    const runningBadge = document.getElementById('runningBadge');
    const statusMsg = document.getElementById('statusMsg');
    const startedAtText = document.getElementById('startedAtText');
    const lastEventAtText = document.getElementById('lastEventAtText');
    const bufferedText = document.getElementById('bufferedText');
    const pausedText = document.getElementById('pausedText');
    const lastLineText = document.getElementById('lastLineText');
    const resultBox = document.getElementById('resultBox');
    const eventCountText = document.getElementById('eventCountText');

    function escapeHtml(input) {
        return String(input || '')
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    function timeText(ts) {
        const stamp = Number(ts || 0);
        if (!stamp) {
            return '-';
        }
        const d = new Date(stamp * 1000);
        if (Number.isNaN(d.getTime())) {
            return '-';
        }
        const y = d.getFullYear();
        const m = String(d.getMonth() + 1).padStart(2, '0');
        const day = String(d.getDate()).padStart(2, '0');
        const hh = String(d.getHours()).padStart(2, '0');
        const mm = String(d.getMinutes()).padStart(2, '0');
        const ss = String(d.getSeconds()).padStart(2, '0');
        return y + '-' + m + '-' + day + ' ' + hh + ':' + mm + ':' + ss;
    }

    function request(url, options) {
        return fetch(url, options).then(async function (response) {
            const data = await response.json().catch(function () {
                return {};
            });
            if (!response.ok) {
                throw new Error(data.error || ('请求失败(' + response.status + ')'));
            }
            return data;
        });
    }

    function normalizeLogLevel(level) {
        const current = String(level || '').trim().toLowerCase();
        if (current === 'warning') {
            return 'warn';
        }
        if (current === 'error') {
            return 'error';
        }
        return current || 'info';
    }

    function buildEventEntry(event) {
        if (!event) {
            return null;
        }
        const rawLine = String(event.line || '').trim();
        const match = rawLine.match(/^\[([^\]]+)\]\s+(INFO|WARN|WARNING|ERROR)\s+(.*)$/i);
        if (match) {
            return {
                level: normalizeLogLevel(match[2]),
                text: '[' + match[1] + '] ' + match[3],
            };
        }
        return {
            level: normalizeLogLevel(event.level),
            text: rawLine || String(event.message || '').trim(),
        };
    }

    function formatEventLineHTML(entry) {
        const raw = String((entry && entry.text) || '');
        const escaped = escapeHtml(raw);
        return escaped.replace(/\b([A-Za-z0-9]+-\d{2,})\b/g, function (match, code) {
            const href = '/movie/' + encodeURIComponent(code);
            return '<a href="' + href + '" class="detail-loop-log-link">' + escapeHtml(code) + '</a>';
        });
    }

    function setStatus(text, type) {
        if (!statusMsg) {
            return;
        }
        statusMsg.textContent = String(text || '');
        statusMsg.className = type === 'err' ? 'small mt-3 err' : (type === 'ok' ? 'small mt-3 ok' : 'small text-muted mt-3');
    }

    function renderEvents() {
        if (!resultBox) {
            return;
        }
        if (!state.events.length) {
            resultBox.textContent = '等待连接...';
            if (eventCountText) {
                eventCountText.textContent = '0 条';
            }
            return;
        }
        resultBox.innerHTML = state.events.map(function (item) {
            const level = normalizeLogLevel(item && item.level);
            return '<span class="event-line log-' + escapeHtml(level) + '">' + formatEventLineHTML(item) + '</span>';
        }).join('');
        resultBox.scrollTop = 0;
        if (eventCountText) {
            eventCountText.textContent = String(state.eventCount) + ' 条';
        }
    }

    function appendEvent(event) {
        const entry = buildEventEntry(event);
        const line = String((entry && entry.text) || '').trim();
        if (!line || !entry) {
            return;
        }
        const key = String((event || {}).at || 0) + '|' + line;
        if (state.seenMap[key]) {
            return;
        }
        const item = {
            key: key,
            text: line,
            level: normalizeLogLevel(entry.level),
        };
        state.seenMap[key] = true;
        state.events.unshift(item);
        if (state.events.length > EVENT_LIMIT) {
            const removed = state.events.pop();
            if (removed) {
                delete state.seenMap[removed.key];
            }
        }
        state.eventCount += 1;
        if (lastEventAtText) {
            lastEventAtText.textContent = timeText(event.at);
        }
        if (lastLineText) {
            lastLineText.textContent = line;
        }
        renderEvents();
    }

    function renderSnapshot(snapshot) {
        const running = !!(snapshot && snapshot.running);
        const paused = !!(snapshot && snapshot.paused);

        if (runningBadge) {
            runningBadge.textContent = running ? (paused ? '已暂停' : '运行中') : '已停止';
            runningBadge.className = 'badge ' + (running ? 'badge-running' : 'badge-stopped');
        }
        if (startedAtText) {
            startedAtText.textContent = timeText(snapshot && snapshot.started_at);
        }
        if (lastEventAtText) {
            lastEventAtText.textContent = timeText(snapshot && snapshot.last_event_at);
        }
        if (bufferedText) {
            bufferedText.textContent = String((snapshot && snapshot.buffered) || 0);
        }
        if (pausedText) {
            pausedText.textContent = paused ? '是' : '否';
        }
        if (lastLineText && snapshot && snapshot.last_line) {
            const entry = buildEventEntry({line: snapshot.last_line, level: 'info'});
            lastLineText.textContent = entry && entry.text ? entry.text : String(snapshot.last_line);
        }
        if (startBtn) {
            startBtn.disabled = running;
        }
        if (stopBtn) {
            stopBtn.disabled = !running;
        }
    }

    function loadSnapshot(silent) {
        request(API_BASE, {method: 'GET'})
            .then(function (snapshot) {
                renderSnapshot(snapshot);
                if (!silent) {
                    setStatus(snapshot && snapshot.running ? '详情抓取 loop 运行中' : '详情抓取 loop 已停止');
                }
            })
            .catch(function (error) {
                if (!silent) {
                    setStatus(error.message, 'err');
                }
            });
    }

    function openStream() {
        if (state.eventSource) {
            state.eventSource.close();
        }
        const source = new EventSource(API_BASE + '/stream');
        state.eventSource = source;
        source.onmessage = function (message) {
            try {
                appendEvent(JSON.parse(message.data));
            } catch (error) {
                setStatus('日志流解析失败', 'err');
            }
        };
        source.onerror = function () {
            if (state.eventSource !== source) {
                return;
            }
            setStatus('日志流连接波动，等待自动重连', 'err');
        };
    }

    function control(action) {
        request(API_BASE + '/' + action, {method: 'POST'})
            .then(function (payload) {
                renderSnapshot(payload.detail_loop || {});
                setStatus(payload.message || '操作完成', 'ok');
            })
            .catch(function (error) {
                setStatus(error.message, 'err');
            });
    }

    if (startBtn) {
        startBtn.addEventListener('click', function () {
            control('start');
        });
    }
    if (stopBtn) {
        stopBtn.addEventListener('click', function () {
            control('stop');
        });
    }

    loadSnapshot(false);
    openStream();
    state.statusTimer = global.setInterval(function () {
        loadSnapshot(true);
    }, 5000);

    global.addEventListener('beforeunload', function () {
        if (state.eventSource) {
            state.eventSource.close();
            state.eventSource = null;
        }
        if (state.statusTimer) {
            global.clearInterval(state.statusTimer);
            state.statusTimer = null;
        }
    });
})(window);
