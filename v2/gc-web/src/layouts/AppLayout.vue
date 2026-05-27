<template>
  <div class="rgc-shell">
    <aside class="rgc-sidebar">
      <div class="rgc-sidebar-head" aria-label="Rudy GC v2">
        <span class="rgc-sidebar-mark"></span>
      </div>
      <nav class="rgc-side-nav" aria-label="主导航">
        <section v-for="group in sidebarGroups" :key="group.group">
          <h2>{{ group.label }}</h2>
          <RouterLink
            v-for="item in group.items"
            :key="item.to"
            :to="item.to"
            class="rgc-side-item"
            :class="{ active: isActive(item.to) }"
          >
            <span class="rgc-side-icon">{{ item.icon }}</span>
            <span>{{ item.label }}</span>
            <span class="rgc-side-arrow">›</span>
          </RouterLink>
        </section>
      </nav>
    </aside>

    <div class="rgc-main">
      <header class="rgc-topbar">
        <button class="rgc-menu-button" type="button" aria-label="菜单">☰</button>
        <nav class="rgc-top-actions" aria-label="快捷入口">
          <RouterLink
            v-for="item in topActions"
            :key="item.path"
            :to="item.path"
            :class="{ active: route.path === item.path }"
          >
            {{ item.title }}
          </RouterLink>
        </nav>
        <span class="rgc-fullscreen" aria-hidden="true">⛶</span>
      </header>
      <main class="rgc-content">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";

import { allNavRoutes, quickNav, type NavRoute } from "@/data/routes";

const route = useRoute();

const groupLabels: Record<NavRoute["group"], string> = {
  Movie: "Movie",
  Library: "Library",
  Ranks: "Ranks",
  People: "People",
  Crawler: "Crawler",
  FetchSite: "Fetch Site",
  Media: "Media",
  Aggregation: "Aggregation",
  SC: "SC",
};

const groupIcons: Record<NavRoute["group"], string> = {
  Movie: "▦",
  Library: "□",
  Ranks: "≡",
  People: "○",
  Crawler: "⌁",
  FetchSite: "↗",
  Media: "▣",
  Aggregation: "∑",
  SC: "◇",
};

const sidebarGroups = computed(() => {
  const groups: Array<{ group: NavRoute["group"]; label: string; items: Array<{ label: string; icon: string; to: string }> }> = [];
  for (const item of allNavRoutes) {
    let group = groups.find((entry) => entry.group === item.group);
    if (!group) {
      group = { group: item.group, label: groupLabels[item.group], items: [] };
      groups.push(group);
    }
    group.items.push({ label: item.title, icon: groupIcons[item.group], to: item.path });
  }
  return groups;
});

const topActions = quickNav;

function isActive(path: string) {
  if (path === "/casts" && route.name === "cast-detail") {
    return true;
  }
  return route.path === path || (path !== "/" && route.path.startsWith(`${path}/`));
}
</script>
