document.addEventListener('click', async function (e) {
    const a = e.target.closest('.open-in-finder');
    if (!a) return;
    e.preventDefault();

    const path = a.getAttribute('data-path') || '';
    if (!path) return;

    try {
        const res = await fetch('/api/open-finder', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ path })
        });

        if (res.ok) {
            // 成功时做个轻提示（变绿一秒）
            a.classList.add('text-success');
            setTimeout(() => a.classList.remove('text-success'), 1200);
        } else {
            const msg = await res.text();
            console.error('打开失败：', msg);
        }
    } catch (err) {
        console.error('网络异常：', err);
    }
})