<template>
  <div class="app app-shell" :class="[routeClass, { 'sidebar-collapsed': sidebarCollapsed }]">
    <a class="skip-link" href="#main-content">跳到页面内容</a>
    <aside class="app-sidebar">
      <div class="sidebar-head">
        <RouterLink class="brand" to="/" aria-label="Rudy GC 首页" :title="sidebarCollapsed ? 'Rudy GC' : undefined">
          <span class="brand-mark"><AppIcon name="photos" /></span>
          <span class="brand-wordmark">Rudy GC<span class="brand-subtitle">MEDIA LIBRARY</span></span>
        </RouterLink>
        <button
          class="sidebar-collapse-button"
          type="button"
          :aria-label="sidebarToggleLabel"
          :title="sidebarToggleLabel"
          :aria-pressed="sidebarCollapsed"
          @click="sidebarCollapsed = !sidebarCollapsed"
        >
          <span aria-hidden="true">{{ sidebarCollapsed ? "›" : "‹" }}</span>
        </button>
      </div>
      <nav ref="navigationRef" class="sidenav" aria-label="主导航">
        <section
          v-for="group in sidebarGroups"
          :key="group.label"
          class="sidenav-group"
          :class="{ active: isGroupActive(group) }"
        >
          <button
            class="sidenav-group-toggle"
            type="button"
            :aria-expanded="isGroupOpen(group)"
            :aria-controls="`sidenav-${group.icon}`"
            :title="sidebarCollapsed ? group.label : undefined"
            @click="toggleGroup(group)"
          >
            <span class="sidenav-group-heading"
              ><AppIcon :name="groupIcons[group.icon] || 'photos'" /><span>{{ group.label }}</span></span
            >
            <span class="sidenav-group-chevron" aria-hidden="true">›</span>
          </button>
          <div
            :id="`sidenav-${group.icon}`"
            class="sidenav-group-panel"
            :class="{ expanded: isGroupOpen(group) }"
            :aria-hidden="isGroupOpen(group) ? undefined : true"
            :inert="!isGroupOpen(group)"
          >
            <div class="sidenav-group-links">
              <RouterLink
                v-for="link in group.links"
                :key="link.href"
                :to="link.href"
                :class="{ 'nav-current': isActive(link.href) }"
                :aria-current="isActive(link.href) ? 'page' : undefined"
                :aria-label="sidebarCollapsed ? link.label : undefined"
                :title="link.label"
                :data-href="link.href"
              >
                <AppIcon :name="groupIcons[group.icon] || 'photos'" /><span>{{ link.label }}</span>
              </RouterLink>
            </div>
          </div>
        </section>
      </nav>
      <span class="library-signature">个人媒体库 · V2</span>
    </aside>
    <div class="app-workspace">
      <header class="workspace-topbar">
        <nav class="workspace-shortcuts" aria-label="快捷导航">
          <section v-for="group in topNavGroups" :key="group.label" class="shortcut-group" :aria-label="group.label">
            <span class="shortcut-group-label">{{ group.label }}</span>
            <RouterLink
              v-for="link in group.links"
              :key="link.href"
              :to="link.href"
              class="strip-link shortcut-link"
              :class="{ active: isTopLinkActive(link.href) }"
              :aria-current="isTopLinkActive(link.href) ? 'page' : undefined"
              >{{ link.label }}</RouterLink
            >
          </section>
        </nav>
        <h1 class="workspace-page-name">{{ currentTitle }}</h1>
      </header>
      <main id="main-content" class="main-content" tabindex="-1"><slot /></main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { sidebarGroups, topNavGroups } from "@/app/navigation";
import AppIcon from "@/shared/components/AppIcon.vue";
import type { NavGroup } from "@/shared/types/gc";

const route = useRoute();
const navigationRef = ref<HTMLElement | null>(null);
const groupIcons: Record<string, string> = {
  MC: "photos",
  RK: "chart",
  AG: "database",
  LS: "albums",
  S: "random",
  TK: "tasks",
  RD: "download",
  FS: "requests",
  SH: "text",
  FC: "photos",
  MD: "archive",
  UI: "palette"
};
const currentTitle = computed(() => String(route.meta.title || "Rudy GC"));
watch(
  currentTitle,
  (title) => {
    document.title = `${title} · Rudy GC V2`;
  },
  { immediate: true }
);
async function revealNavigation() {
  await nextTick();
  const nav = navigationRef.value;
  const active = nav?.querySelector<HTMLElement>(".nav-current");
  if (!nav || !active) return;
  const navRect = nav.getBoundingClientRect();
  const activeRect = active.getBoundingClientRect();
  if (activeRect.top < navRect.top) nav.scrollTop -= navRect.top - activeRect.top;
  else if (activeRect.bottom > navRect.bottom) nav.scrollTop += activeRect.bottom - navRect.bottom;
}
const sidebarCollapsed = ref(false);
const expandedGroupLabels = ref<string[]>([]);
const collapsedGroupLabels = ref<string[]>([]);
const legacyListCardRouteClasses: Record<string, string> = {
  "/fclistcard": "route-cards",
  "/fclistcardowned": "route-cardsmediamowned",
  "/fclistcardrandom": "route-cardsrandom",
  "/listcard": "route-cards",
  "/listcardtoday": "route-cardstoday",
  "/listcardrank": "route-cardshasrank",
  "/listcardowned": "route-cardsmediamowned",
  "/listcardneed": "route-cardsneeddownload",
  "/listcardrandom": "route-cardsrandom",
  "/listcardperiodbirth": "route-cardsperiod",
  "/listcardperiodrelease": "route-cardsperiod",
  "/listcardagg": "route-cardsagg",
  "/listcarddayrank": "route-moviecarddayrank",
  "/listcardperiodrank": "route-moviecardperiodrank"
};
const legacyRenamedRouteClasses: Record<string, string> = {
  "/dailybestfetch": "route-triggers-dailybest",
  "/seedfetch": "route-triggers-seeds",
  "/postprocess": "route-triggers-post-process",
  "/databackfill": "route-triggers-backfill",
  "/sitefetch": "route-triggers-fetch-site",
  "/javbusfetch": "route-triggers-fetch-site-javbus-filtered",
  "/sukebeifetch": "route-triggers-fetch-site-sukebei-filtered",
  "/scadd": "route-triggers-sc",
  "/scmove": "route-triggers-sc-media",
  "/mediaingest": "route-triggers-media",
  "/fcingest": "route-triggers-fc2-media",
  "/mediarescan": "route-triggers-media-rescan",
  "/mediarollback": "route-triggers-media-rollback",
  "/scevents": "route-sc-events",
  "/sccards": "route-sc-events-cards",
  "/rankdaystats": "route-sc-events route-rank-day-stats",
  "/rankperiodstats": "route-sc-events route-rank-period-stats",
  "/scoptionlist": "route-records route-sc-options",
  "/crawlrecordlist": "route-records route-crawl-records route-crawlrecords",
  "/crawltasklist": "route-crawler-tasks route-crawltasks",
  "/detailfetch": "route-crawler-detail-loop",
  "/siteconfiglist": "route-fetch-sites route-siteconfig",
  "/javbuslist": "route-fetch-site-javbus-list",
  "/sukebeilist": "route-fetch-site-sukebei-list",
  "/shbucketlist": "route-bucketssh route-shbuckets",
  "/medialist": "route-medias",
  "/aggeventlist": "route-w-agg-events route-wevents",
  "/torrentalbumlist": "route-torrent-albums route-torrentalbums",
  "/moviealbumlist": "route-movie-albums route-moviealbums",
  "/seedlist": "route-d-seeds route-seeddata",
  "/castlist": "route-casts",
  "/eitemlist": "route-e-items route-eitems"
};

const sidebarToggleLabel = computed(() => (sidebarCollapsed.value ? "展开侧边导航" : "收起侧边导航"));
const activePath = computed(() => {
  const detailParents: [string, string][] = [
    ["/moviedetail/", "/listcardtoday"],
    ["/castdetail/", "/castlist"],
    ["/fcdetail/", "/fclist"],
    ["/shthread/", "/shpool"],
    ["/scevent/", "/scevents"],
    ["/remotedailybestfetchrundetail/", "/remotedailybestfetchrunlist"]
  ];
  return detailParents.find(([prefix]) => route.path.startsWith(prefix))?.[1] || route.path;
});
const routeClass = computed(() => {
  const slug = route.path.replace(/^\/+/, "").replace(/[^a-zA-Z0-9_-]+/g, "-") || "home";
  return [
    legacyListCardRouteClasses[route.path] || "",
    legacyRenamedRouteClasses[route.path] || "",
    route.path.startsWith("/moviedetail/") ? "route-movie-detail" : "",
    route.path.startsWith("/castdetail/") ? "route-cast-detail" : "",
    route.path === "/shfetchall" ? "route-triggers-fetch-sehuatang-sh" : "",
    route.path === "/shfetch" ? "route-triggers-fetch-sehuatang-sh-thread-list" : "",
    route.path === "/shfill" ? "route-triggers-fetch-sehuatang-sh-thread-detail" : "",
    route.path === "/shpool" ? "route-fetch-site-sehuatang-sh-list" : "",
    route.path.startsWith("/shthread/") ? "route-fetch-site-sehuatang-sh-detail" : "",
    route.path === "/shmagnetlist" ? "route-fetch-site-sehuatang-sh-list route-sh-magnets route-shmagnets" : "",
    route.path === "/fclist" || route.path.startsWith("/fcdetail/") ? "route-fc2" : "",
    route.path === "/fclist" ? "route-fc2-list" : "",
    route.path.startsWith("/fcdetail/") ? "route-fc2-detail" : "",
    route.path.startsWith("/scevent/") ? "route-sc-events" : "",
    route.path.startsWith("/remotedailybest") ? "route-remote-dailybest" : "",
    `route-${slug}`
  ]
    .filter(Boolean)
    .join(" ");
});

function isActive(href: string): boolean {
  if (href === "/") {
    return activePath.value === "/";
  }
  return activePath.value === href || activePath.value.startsWith(`${href}/`);
}

function isTopLinkActive(href: string): boolean {
  return isActive(href);
}

function isGroupOpen(group: NavGroup): boolean {
  if (collapsedGroupLabels.value.includes(group.label)) {
    return false;
  }
  return expandedGroupLabels.value.includes(group.label) || isGroupActive(group);
}

function isGroupActive(group: NavGroup): boolean {
  return group.links.some((link) => isActive(link.href));
}

function toggleGroup(group: NavGroup): void {
  const label = group.label;
  if (isGroupOpen(group)) {
    expandedGroupLabels.value = expandedGroupLabels.value.filter((item) => item !== label);
    collapsedGroupLabels.value = collapsedGroupLabels.value.includes(label)
      ? collapsedGroupLabels.value
      : [...collapsedGroupLabels.value, label];
    return;
  }

  collapsedGroupLabels.value = collapsedGroupLabels.value.filter((item) => item !== label);
  expandedGroupLabels.value = expandedGroupLabels.value.includes(label)
    ? expandedGroupLabels.value
    : [...expandedGroupLabels.value, label];
}

watch(
  () => route.path,
  () => {
    collapsedGroupLabels.value = [];
    void revealNavigation();
  }
);
</script>
