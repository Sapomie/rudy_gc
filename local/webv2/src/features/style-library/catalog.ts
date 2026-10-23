export const styleLibraryPages = [
  { path: "/styles", label: "页面目录", title: "样式库", description: "集中查看页面与组件。", icon: "palette" },
  {
    path: "/styles/base",
    label: "基础",
    title: "颜色、排版与图标",
    description: "粉色主题、语义色、文字层级、边框与间距。",
    icon: "palette"
  },
  {
    path: "/styles/gallery",
    label: "图库",
    title: "卡片与相框",
    description: "Movie / FC2 卡片、封面、角标与播放状态。",
    icon: "photos"
  },
  {
    path: "/styles/table",
    label: "表格",
    title: "紧凑数据列表",
    description: "30px 行高、排序、多选、分页与状态区分。",
    icon: "database"
  },
  {
    path: "/styles/form",
    label: "表单",
    title: "按钮与输入控件",
    description: "操作层级、按钮组、筛选与表单反馈。",
    icon: "capture"
  },
  {
    path: "/styles/detail",
    label: "详情",
    title: "封面与信息面板",
    description: "完整封面、信息字段、页签与预览弹层。",
    icon: "text"
  },
  {
    path: "/styles/feedback",
    label: "反馈",
    title: "状态与任务反馈",
    description: "加载、空状态、错误、进度与确认弹层。",
    icon: "tasks"
  }
];

export const semanticStates = [
  { label: "已完成", tone: "good" },
  { label: "下载中", tone: "info" },
  { label: "本地缺失", tone: "warn" },
  { label: "执行失败", tone: "danger" },
  { label: "未开始", tone: "neutral" }
];
