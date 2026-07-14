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
      { href: "/medialist", label: "Medias" },
      { href: "/aggeventlist", label: "WAggEvents" },
      { href: "/torrentalbumlist", label: "TorrentAlbums" },
      { href: "/moviealbumlist", label: "MovieAlbums" },
      { href: "/seedlist", label: "DSeeds" },
      { href: "/castlist", label: "Casts" },
      { href: "/eitemlist", label: "EItems" }
    ]
  },
  {
    label: "SC",
    icon: "S",
    links: [
      { href: "/scevents", label: "ScEvents" },
      { href: "/scpick", label: "ScPickSmartMedia" },
      { href: "/scadd", label: "ScPreview" },
      { href: "/scmove", label: "ScMediaMove" }
    ]
  },
  {
    label: "抓取任务",
    icon: "TK",
    links: [
      { href: "/crawlrecordlist", label: "CrawlRecords" },
      { href: "/dailybestfetch", label: "DailyBest" },
      { href: "/seedfetch", label: "Seeds" },
      { href: "/postprocess", label: "PostProcess" },
      { href: "/databackfill", label: "Backfill" },
      { href: "/detailfetch", label: "DetailLoop" },
      { href: "/crawltasklist", label: "CrawlerTasks" }
    ]
  },
  {
    label: "FetchSite操作",
    icon: "FS",
    links: [
      { href: "/sitefetch", label: "FetchSite" },
      { href: "/sehuatangfetch", label: "FetchSehuatang" },
      { href: "/siteconfiglist", label: "FetchSitesConfig" },
      { href: "/javbuslist", label: "JavbusList" },
      { href: "/sukebeilist", label: "SukebeiList" },
      { href: "/sehuatanglist", label: "SehuatangList" }
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
      { href: "/shmagnetlist", label: "SH磁力" },
      { href: "/shbucketlist", label: "SH聚合桶" }
    ]
  },
  {
    label: "FC2",
    icon: "FC",
    links: [
      { href: "/fclist", label: "FC2影片" },
      { href: "/fclistcard", label: "FC2Cards" },
      { href: "/fclistcardowned", label: "FC2Owned" },
      { href: "/fclistcardrandom", label: "FC2Random" },
      { href: "/fcingest", label: "FC2Media" }
    ]
  },
  {
    label: "Media操作",
    icon: "MD",
    links: [
      { href: "/mediaingest", label: "Media" },
      { href: "/mediarescan", label: "MediaRescan" },
      { href: "/mediarollback", label: "MediaRollback" }
    ]
  }
];

export const topNavGroups: { label: string; links: NavLink[] }[] = [
  {
    label: "常用",
    links: [
      { href: "/listcardtoday", label: "今日卡片" },
      { href: "/crawltasklist", label: "任务列表" },
      { href: "/dailybestfetch", label: "每日榜" }
    ]
  },
  {
    label: "片单",
    links: [
      { href: "/moviealbumlist", label: "影片专辑" },
      { href: "/torrentalbumlist", label: "种子专辑" }
    ]
  },
  {
    label: "抓取",
    links: [
      { href: "/sehuatangfetch", label: "色花堂" },
      { href: "/shfetchall", label: "SH一体" }
    ]
  },
  {
    label: "色花堂SH",
    links: [
      { href: "/shfetch", label: "列表建池" },
      { href: "/shfill", label: "详情补全" },
      { href: "/shpool", label: "帖子池" },
      { href: "/shmagnetlist", label: "SH磁力" },
      { href: "/fclist", label: "FC2影片" },
      { href: "/shbucketlist", label: "聚合桶" }
    ]
  }
];
