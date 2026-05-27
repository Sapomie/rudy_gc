<template>
  <div class="rgc-shell">
    <aside class="rgc-sidebar">
      <div class="rgc-sidebar-head" aria-label="Rudy GC v2">
        <span class="rgc-sidebar-mark"></span>
      </div>
      <nav class="rgc-side-nav" aria-label="主导航">
        <section>
          <h2>Main</h2>
          <RouterLink
            v-for="item in sidebarItems"
            :key="item.label"
            :to="item.to"
            class="rgc-side-item"
            :class="{ active: route.path === item.to }"
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
import { useRoute } from "vue-router";

import { cardRoutes } from "@/data/routes";

const route = useRoute();

const sidebarItems = cardRoutes.map((item) => ({
  label: item.key === "cards" ? "MovieCard" : `MovieCard ${item.title}`,
  icon: "▦",
  to: item.path,
}));

const topActions = cardRoutes;
</script>
