// 样式库专用示例；不读取或修改业务数据。
export function coverArtwork(index = 0): string {
  const palettes = [
    ["#f5dce7", "#be4b78", "#6c2947"],
    ["#dce8f8", "#4976b5", "#203958"],
    ["#e6e0f4", "#8b73ab", "#4a3b60"],
    ["#dfebe2", "#6f9278", "#354d3b"]
  ];
  const [light, mid, dark] = palettes[index % palettes.length];
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="600" height="900" viewBox="0 0 600 900"><defs><linearGradient id="g" x2="0.8" y2="1"><stop stop-color="${light}"/><stop offset="1" stop-color="${mid}"/></linearGradient></defs><rect width="600" height="900" fill="url(#g)"/><circle cx="430" cy="240" r="220" fill="${light}" opacity=".6"/><path d="M-80 800L320 160L730 900H-80Z" fill="${dark}" opacity=".22"/><path d="M-80 900L130 410L510 900Z" fill="${dark}" opacity=".24"/><rect x="38" y="38" width="524" height="824" rx="250" fill="none" stroke="${light}" stroke-width="2" opacity=".7"/><text x="52" y="728" fill="${light}" font-family="Arial,sans-serif" font-size="18" letter-spacing="6">RUDY GC / COLLECTION</text><text x="48" y="800" fill="white" font-family="Arial,sans-serif" font-size="58" letter-spacing="3">FRAME ${String(index + 1).padStart(2, "0")}</text></svg>`;
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;
}

export const galleryCards = Array.from({ length: 12 }, (_, index) => {
  const fc2 = index >= 6;
  const name = fc2 ? `FC2-${100001 + index}` : `DEMO-${String(index + 1).padStart(3, "0")}`;
  const title = [
    "有字幕 · 双角标",
    "无字幕 · 蓝色播放",
    "单个角标",
    "无角标 · 普通封面",
    "暂无封面",
    "已移除 · 不可播放"
  ][index % 6];
  const cover = index % 6 === 4 ? "" : coverArtwork(index);
  const come = index % 6 < 3 ? 2 : 0;
  const sc = index % 6 < 2 ? 3 : 0;
  return {
    id: index + 1,
    kind: fc2 ? "fc2" : "movie",
    name,
    title,
    movie: {
      name,
      title,
      prefix: "DEMO",
      jacket_img: cover,
      score: "8.1",
      viewers_number_watched: 128,
      owned_w_media: index % 6 === 5 ? 6 : index % 2 ? 5 : 4,
      video_url_w_media: "#preview",
      film_birth_date_w_media: "2026-10-23",
      releasing_date: "2026-10-01",
      come_times: come,
      sc_times: sc,
      highest_rank: index === 0 ? 12 : 0,
      genre: ["示例分类", "收藏"],
      cast: [{ Name: "示例人物" }],
      director: "示例作者",
      bus_url: "#source",
      FC2Key: name,
      Title: title,
      CoverImageURL: cover,
      PerformerText: "示例人物",
      ShowPlayButton: index % 6 !== 5,
      VideoURL: "#preview",
      PlayButtonClass: index % 2 ? "btn-info" : "btn-success",
      PlayButtonText: "Play",
      ComeTimes: come,
      ScTimes: sc,
      DownloadTime: 1792684800,
      PostTime: 1790784000
    }
  };
});

export const tableRows = Array.from({ length: 18 }, (_, index) => ({
  id: index + 1,
  name: `DEMO-${String(index + 1).padStart(3, "0")}`,
  title: ["每日统计", "媒体入库", "封面补全", "列表抓取", "详情更新", "相册整理"][index % 6],
  date: `2026-10-${String(23 - index).padStart(2, "0")} 09:30`,
  count: 120 + index * 7,
  status: index % 5
}));
