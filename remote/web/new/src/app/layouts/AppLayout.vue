<template>
  <div class="app-shell" :class="[routeClass, { 'is-sidebar-collapsed': sidebarCollapsed }]">
    <aside class="side-panel">
      <RouterLink class="brand-block" to="/dailybestoverview">
        <span class="brand-mark">RG</span>
        <span class="brand-text">Rudy-GC Remote</span>
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
            :class="{ active: isActive(link.href) }"
            :aria-current="isActive(link.href) ? 'page' : undefined"
            :title="link.label"
            :to="link.href"
          >
            {{ link.label }}
          </RouterLink>
        </section>
      </nav>
    </header>

    <main class="main-content">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { sidebarGroups, topNavGroups, type NavGroup } from "@/app/navigation";

const route = useRoute();
const sidebarCollapsed = ref(false);
const expandedGroupLabels = ref<string[]>([]);
const collapsedGroupLabels = ref<string[]>([]);
const currentTitle = computed(() => String(route.meta.title || "Remote 管理端"));
const sidebarToggleLabel = computed(() => (sidebarCollapsed.value ? "展开侧边导航" : "收起侧边导航"));
const routeClass = computed(() => {
  const slug = route.path.replace(/^\/+/, "").replace(/[^a-zA-Z0-9_-]+/g, "-") || "overview";
  return `route-${slug}`;
});

function isActive(href: string): boolean {
  return route.path === href || route.path.startsWith(`${href}/`);
}

function isGroupActive(group: NavGroup): boolean {
  return group.links.some((link) => isActive(link.href));
}

function isGroupOpen(group: NavGroup): boolean {
  if (collapsedGroupLabels.value.includes(group.label)) return false;
  return expandedGroupLabels.value.includes(group.label) || isGroupActive(group);
}

function toggleGroup(group: NavGroup): void {
  if (isGroupOpen(group)) {
    expandedGroupLabels.value = expandedGroupLabels.value.filter((label) => label !== group.label);
    if (!collapsedGroupLabels.value.includes(group.label)) {
      collapsedGroupLabels.value = [...collapsedGroupLabels.value, group.label];
    }
    return;
  }
  collapsedGroupLabels.value = collapsedGroupLabels.value.filter((label) => label !== group.label);
  if (!expandedGroupLabels.value.includes(group.label)) {
    expandedGroupLabels.value = [...expandedGroupLabels.value, group.label];
  }
}

watch(
  () => route.path,
  () => {
    collapsedGroupLabels.value = [];
  }
);
</script>
