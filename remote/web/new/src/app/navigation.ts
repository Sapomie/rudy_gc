export interface NavLink {
  href: string;
  label: string;
}

export interface NavGroup {
  label: string;
  icon: string;
  links: NavLink[];
}

export const sidebarGroups: NavGroup[] = [
  {
    label: "数据管理",
    icon: "DB",
    links: [
      { href: "/dailybestoverview", label: "运行概览" },
      { href: "/ranklist", label: "排名明细" },
      { href: "/rankdaystats", label: "每日统计" },
      { href: "/rankperiodlist", label: "周期排行" },
      { href: "/rankperioditemlist", label: "周期明细" },
      { href: "/rankperiodstats", label: "周期统计" }
    ]
  },
  {
    label: "采集管理",
    icon: "CL",
    links: [
      { href: "/dailybestfetchrunlist", label: "抓取任务" },
      { href: "/dailybestfetch", label: "手动抓取" }
    ]
  }
];

export const topNavGroups: NavGroup[] = [
  {
    label: "数据",
    icon: "DB",
    links: [
      { href: "/dailybestoverview", label: "运行概览" },
      { href: "/ranklist", label: "排名明细" },
      { href: "/rankdaystats", label: "每日统计" },
      { href: "/rankperiodlist", label: "周期排行" },
      { href: "/rankperioditemlist", label: "周期明细" },
      { href: "/rankperiodstats", label: "周期统计" }
    ]
  },
  {
    label: "采集",
    icon: "CL",
    links: [
      { href: "/dailybestfetchrunlist", label: "抓取任务" },
      { href: "/dailybestfetch", label: "手动抓取" }
    ]
  }
];
