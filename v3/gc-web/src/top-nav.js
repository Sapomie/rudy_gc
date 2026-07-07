const navGroups = [
  {
    label: "常用",
    links: [
      {href: "/crawler/tasks", label: "任务列表"},
      {href: "/movie-albums", label: "MovieAlbums"},
      {href: "/torrent-albums", label: "TorrentAlbums"},
      {href: "/triggers/dailybest", label: "DailyBest"},
      {href: "/triggers/fetch-sehuatang", label: "FetchSehuatang"}
    ]
  },
  {
    label: "色花堂SH",
    links: [
      {href: "/triggers/fetch-sehuatang-sh", label: "SH一体抓取"},
      {href: "/triggers/fetch-sehuatang-sh-thread-list", label: "SH列表建池"},
      {href: "/triggers/fetch-sehuatang-sh-thread-detail", label: "SH详情补全"},
      {href: "/fetch-site-sehuatang-sh-list", label: "SH帖子池"}
    ]
  }
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
  root.innerHTML = navGroups.map((group, index) => {
    const spacing = index === navGroups.length - 1 ? "" : " me-2";
    const links = group.links.map((item) =>
      `<a href="${escapeHtml(item.href)}" class="btn btn-sm btn-outline-primary">${escapeHtml(item.label)}</a>`
    ).join("");
    return `<div class="btn-group btn-group-sm${spacing}" role="group" aria-label="${escapeHtml(group.label)}">
      <span class="btn btn-sm btn-light disabled">${escapeHtml(group.label)}</span>
      ${links}
    </div>`;
  }).join("");
}

renderTopNav();
