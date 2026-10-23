import type { RouteRecordRaw } from "vue-router";

export const styleLibraryRoutes: RouteRecordRaw[] = [
  {
    path: "/styles",
    component: () => import("./pages/StyleLibraryPage.vue"),
    children: [
      {
        path: "",
        name: "style-library",
        component: () => import("./pages/OverviewPage.vue"),
        meta: { title: "样式库" }
      },
      {
        path: "base",
        name: "style-foundation",
        component: () => import("./pages/FoundationPage.vue"),
        meta: { title: "样式库 · 基础" }
      },
      {
        path: "gallery",
        name: "style-gallery",
        component: () => import("./pages/GalleryPage.vue"),
        meta: { title: "样式库 · 图库" }
      },
      {
        path: "table",
        name: "style-table",
        component: () => import("./pages/TablePage.vue"),
        meta: { title: "样式库 · 表格" }
      },
      {
        path: "form",
        name: "style-form",
        component: () => import("./pages/FormPage.vue"),
        meta: { title: "样式库 · 表单" }
      },
      {
        path: "detail",
        name: "style-detail",
        component: () => import("./pages/DetailPage.vue"),
        meta: { title: "样式库 · 详情" }
      },
      {
        path: "feedback",
        name: "style-feedback",
        component: () => import("./pages/FeedbackPage.vue"),
        meta: { title: "样式库 · 反馈" }
      }
    ]
  }
];
