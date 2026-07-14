<template>
  <div class="app-shell" :class="[routeClass, { 'is-sidebar-collapsed': sidebarCollapsed }]">
    <aside class="side-panel">
      <RouterLink class="brand-block" to="/">
        <span class="brand-mark">R2</span>
        <span class="brand-text">Rudy-GC</span>
      </RouterLink>

      <nav class="side-nav" aria-label="主导航">
        <section
          v-for="group in sidebarGroups"
          :key="group.label"
          class="side-group"
          :class="{ 'is-open': isGroupOpen(group), 'is-active-group': isGroupActive(group) }"
        >
          <button
            class="side-group-title"
            type="button"
            :aria-expanded="isGroupOpen(group)"
            @click="toggleGroup(group)"
          >
            <span class="side-label">{{ group.label }}</span>
            <span class="side-count" aria-hidden="true">{{ group.links.length }}</span>
          </button>
          <RouterLink
            v-for="link in group.links"
            :key="link.href"
            class="side-link"
            :class="{ active: isActive(link.href) }"
            :aria-current="isActive(link.href) ? 'page' : undefined"
            :data-href="link.href"
            :title="link.label"
            :to="link.href"
          >
            {{ link.label }}
          </RouterLink>
        </section>
      </nav>
    </aside>

    <header class="top-bar app-topbar">
      <div class="top-left top-title-zone">
        <button
          class="icon-button top-menu-button"
          type="button"
          :aria-label="sidebarToggleLabel"
          :aria-pressed="sidebarCollapsed"
          @click="sidebarCollapsed = !sidebarCollapsed"
        >
          ☰
        </button>
        <span class="top-title">{{ currentTitle }}</span>
      </div>
      <nav class="top-links top-nav-strip" aria-label="快捷导航">
        <section v-for="group in topNavGroups" :key="group.label" class="top-link-group top-nav-section">
          <span class="top-link-label top-nav-label">{{ group.label }}</span>
          <RouterLink
            v-for="link in group.links"
            :key="link.href"
            class="top-link top-nav-item"
            :class="{ active: isTopLinkActive(link.href) }"
            :aria-current="isTopLinkActive(link.href) ? 'page' : undefined"
            :title="link.label"
            :to="link.href"
          >
            {{ link.label }}
          </RouterLink>
        </section>
      </nav>
      <div v-if="showLegacyTopActions" class="top-legacy-actions top-action-zone" aria-label="页面操作">
        <button class="top-legacy-button fullscreen" type="button" aria-label="全屏">⛶</button>
        <button class="top-legacy-button user" type="button" aria-label="用户"></button>
      </div>
    </header>

    <main class="main-content">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { sidebarGroups, topNavGroups } from "@/data/navigation";
import type { NavGroup } from "@/types/gc";
import { routeTitle } from "@/utils/format";

const route = useRoute();
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
  "/scpick": "route-sc-pick-smart-media",
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

const currentTitle = computed(() => String(route.meta.title || routeTitle(route.path)));
const sidebarToggleLabel = computed(() => (sidebarCollapsed.value ? "展开侧边导航" : "收起侧边导航"));
const activePath = computed(() => route.path);
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
    `route-${slug}`
  ].filter(Boolean).join(" ");
});
const showLegacyTopActions = computed(() => (
  route.path === "/listcarddayrank"
  || route.path === "/listcardperiodrank"
  || route.path.startsWith("/moviedetail/")
  || route.path.startsWith("/castdetail/")
));

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

watch(() => route.path, () => {
  collapsedGroupLabels.value = [];
});
</script>
