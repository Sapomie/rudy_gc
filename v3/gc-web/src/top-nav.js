const links = [
  {href: "/crawler/tasks", label: "任务列表"},
  {href: "/movie-albums", label: "MovieAlbums"},
  {href: "/torrent-albums", label: "TorrentAlbums"},
  {href: "/triggers/dailybest", label: "DailyBest"},
  {href: "/triggers/fetch-sehuatang", label: "FetchSehuatang"}
];

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function renderTopNav() {
  const root = document.getElementById("top-nav-root");
  if (!root) {
    return;
  }
  root.innerHTML = links.map((item, index) => {
    const spacing = index === links.length - 1 ? "" : " me-2";
    return `<a href="${escapeHtml(item.href)}" class="btn btn-sm${spacing} btn-outline-primary">${escapeHtml(item.label)}</a>`;
  }).join("");
}

renderTopNav();
