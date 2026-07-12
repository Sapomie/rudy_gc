import type { NavGroup, NavLink } from "@/types/gc";

export const sidebarGroups: NavGroup[] = [
  {
    label: "MovieCard",
    icon: "MC",
    links: [
      { href: "/listcardtoday", label: "Today" },
      { href: "/listcardowned", label: "MediaOwned" },
      { href: "/listcardrank", label: "HasRank" },
      { href: "/listcarddayrank", label: "DayRank" },
      { href: "/listcardperiodrank", label: "PeriodRank" },
      { href: "/listcardneed", label: "NeedDownload" },
      { href: "/listcardrandom", label: "Random" },
      { href: "/listcardagg", label: "Agg" }
    ]
  },
  {
    label: "MovieAgg",
    icon: "AG",
    links: [
      { href: "/aggbackfill", label: "AggBackfill" },
      { href: "/mediadir", label: "MediaDirectory" },
      { href: "/aggbirth", label: "Agg-WMedia-Birth" },
      { href: "/bucketsbirth", label: "Agg-WMedia-Bucket" },
      { href: "/aggrelease", label: "Agg-All-Release" },
      { href: "/bucketsrelease", label: "Agg-All-Release-Bucket" }
    ]
  },
  {
    label: "List",
    icon: "LS",
    links: [
      { href: "/medias", label: "Medias" },
      { href: "/w-agg-events", label: "WAggEvents" },
      { href: "/torrent-albums", label: "TorrentAlbums" },
      { href: "/movie-albums", label: "MovieAlbums" },
      { href: "/d-seeds", label: "DSeeds" },
      { href: "/casts", label: "Casts" },
      { href: "/e-items", label: "EItems" }
    ]
  },
  {
    label: "SC",
    icon: "S",
    links: [
      { href: "/sc-events", label: "ScEvents" },
      { href: "/sc-pick-smart-media", label: "ScPickSmartMedia" },
      { href: "/triggers/sc", label: "ScPreview" },
      { href: "/triggers/sc-media", label: "ScMediaMove" }
    ]
  },
  {
    label: "抓取任务",
    icon: "TK",
    links: [
      { href: "/crawl-records", label: "CrawlRecords" },
      { href: "/triggers/dailybest", label: "DailyBest" },
      { href: "/triggers/seeds", label: "Seeds" },
      { href: "/triggers/post-process", label: "PostProcess" },
      { href: "/triggers/backfill", label: "Backfill" },
      { href: "/crawler/detail-loop", label: "DetailLoop" },
      { href: "/crawler/tasks", label: "CrawlerTasks" }
    ]
  },
  {
    label: "FetchSite操作",
    icon: "FS",
    links: [
      { href: "/triggers/fetch-site", label: "FetchSite" },
      { href: "/triggers/fetch-sehuatang", label: "FetchSehuatang" },
      { href: "/fetch-sites", label: "FetchSitesConfig" },
      { href: "/fetch-site-javbus-list", label: "JavbusList" },
      { href: "/fetch-site-sukebei-list", label: "SukebeiList" },
      { href: "/fetch-site-sehuatang-list", label: "SehuatangList" }
    ]
  },
  {
    label: "色花堂SH",
    icon: "SH",
    links: [
      { href: "/shfetchall", label: "SH一体抓取" },
      { href: "/shfetch", label: "SH列表建池" },
      { href: "/shfill", label: "SH详情补全" },
      { href: "/shpool", label: "SH帖子池" },
      { href: "/bucketssh", label: "SH聚合桶" }
    ]
  },
  {
    label: "Media操作",
    icon: "MD",
    links: [
      { href: "/triggers/media", label: "Media" },
      { href: "/triggers/media-rescan", label: "MediaRescan" },
      { href: "/triggers/media-rollback", label: "MediaRollback" }
    ]
  }
];

export const topNavGroups: { label: string; links: NavLink[] }[] = [
  {
    label: "常用",
    links: [
      { href: "/listcardtoday", label: "今日卡片" },
      { href: "/crawler/tasks", label: "任务列表" },
      { href: "/triggers/dailybest", label: "每日榜" }
    ]
  },
  {
    label: "片单",
    links: [
      { href: "/movie-albums", label: "影片专辑" },
      { href: "/torrent-albums", label: "种子专辑" }
    ]
  },
  {
    label: "抓取",
    links: [
      { href: "/triggers/fetch-sehuatang", label: "色花堂" },
      { href: "/shfetchall", label: "SH一体" }
    ]
  },
  {
    label: "色花堂SH",
    links: [
      { href: "/shfetch", label: "列表建池" },
      { href: "/shfill", label: "详情补全" },
      { href: "/shpool", label: "帖子池" },
      { href: "/bucketssh", label: "聚合桶" }
    ]
  }
];
