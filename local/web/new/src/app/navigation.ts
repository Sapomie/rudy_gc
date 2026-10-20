import type { NavGroup, NavLink } from "@/shared/types/gc";

export const sidebarGroups: NavGroup[] = [
  {
    label: "MovieCard",
    icon: "MC",
    links: [
      { href: "/listcardtoday", label: "Today" },
      { href: "/listcardperiodbirth", label: "下载周期卡片" },
      { href: "/listcardperiodrelease", label: "上映周期卡片" },
      { href: "/listcardowned", label: "MediaOwned" },
      { href: "/listcardneed", label: "NeedDownload" },
      { href: "/listcardrandom", label: "Random" }
    ]
  },
  {
    label: "Rank",
    icon: "RK",
    links: [
      { href: "/listcardrank", label: "HasRank" },
      { href: "/listcarddayrank", label: "DayRank" },
      { href: "/listcardperiodrank", label: "PeriodRank" },
      { href: "/rankdaystats", label: "DayStats" },
      { href: "/rankperiodstats", label: "PeriodStats" }
    ]
  },
  {
    label: "MovieAgg",
    icon: "AG",
    links: [
      { href: "/aggbackfill", label: "AggBackfill" },
      { href: "/mediadir", label: "MediaDirectory" },
      { href: "/bucketsbirth", label: "Agg-WMedia-Bucket" },
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
      { href: "/eitemlist", label: "EItems" },
      { href: "/listcardagg", label: "Agg" },
      { href: "/crawlrecordlist", label: "CrawlRecords" }
    ]
  },
  {
    label: "SC",
    icon: "S",
    links: [
      { href: "/scevents", label: "ScEvents" },
      { href: "/scoptionlist", label: "ScOptions" },
      { href: "/scpick", label: "ScPick" },
      { href: "/scadd", label: "ScPreview" },
      { href: "/scmove", label: "ScMediaMove" }
    ]
  },
  {
    label: "抓取任务",
    icon: "TK",
    links: [
      { href: "/dailybestfetch", label: "DailyBest" },
      { href: "/seedfetch", label: "Seeds" },
      { href: "/postprocess", label: "PostProcess" },
      { href: "/databackfill", label: "Backfill" },
      { href: "/detailfetch", label: "DetailLoop" },
      { href: "/crawltasklist", label: "CrawlerTasks" }
    ]
  },
  {
    label: "远端 DailyBest",
    icon: "RD",
    links: [
      { href: "/remotedailybestrankdaystats", label: "每日统计" },
      { href: "/remotedailybestfetchrunlist", label: "抓取任务" },
      { href: "/remotedailybestfetchloglist", label: "请求日志" },
      { href: "/remotedailybesttargetprobelist", label: "目标探测" },
      { href: "/remotedailybestfetch", label: "手动抓取" },
      { href: "/remotedailybestsourceconfig", label: "站点配置" }
    ]
  },
  {
    label: "FetchSite操作",
    icon: "FS",
    links: [
      { href: "/sitefetch", label: "FetchSite" },
      { href: "/siteconfiglist", label: "FetchSitesConfig" },
      { href: "/javbuslist", label: "JavbusList" },
      { href: "/sukebeilist", label: "SukebeiList" }
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
      { href: "/fclistcardrandom", label: "FC2Random" }
    ]
  },
  {
    label: "Media操作",
    icon: "MD",
    links: [
      { href: "/mediaingest", label: "Media" },
      { href: "/fcingest", label: "FC2Media" },
      { href: "/mediarescan", label: "MediaRescan" },
      { href: "/mediarollback", label: "MediaRollback" }
    ]
  }
];

export const topNavGroups: { label: string; links: NavLink[] }[] = [
  {
    label: "统计",
    links: [
      { href: "/remotedailybestrankdaystats", label: "每日统计" },
      { href: "/rankdaystats", label: "日榜统计" }
    ]
  },
  {
    label: "色花堂",
    links: [
      { href: "/shfetchall", label: "色花堂一体抓取" },
      { href: "/shmagnetlist", label: "色花堂磁力" }
    ]
  },
  {
    label: "任务记录",
    links: [
      { href: "/crawltasklist", label: "任务列表" },
      { href: "/scevents", label: "SC事件" }
    ]
  },
  {
    label: "片单",
    links: [
      { href: "/moviealbumlist", label: "影片专辑" },
      { href: "/torrentalbumlist", label: "种子专辑" }
    ]
  }
];
