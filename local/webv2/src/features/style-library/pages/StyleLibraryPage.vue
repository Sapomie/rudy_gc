<template>
  <PageShell title="样式库" class="style-library-page">
    <nav class="style-library-index" aria-label="样式库分类">
      <RouterLink
        v-for="page in styleLibraryPages"
        :key="page.path"
        :to="page.path"
        :class="{ active: route.path === page.path }"
        :aria-current="route.path === page.path ? 'page' : undefined"
      >
        <AppIcon :name="page.icon" /><span>{{ page.label }}</span>
      </RouterLink>
      <span class="style-library-index-note">Rudy GC · 粉色主题</span>
    </nav>
    <RouterView />
  </PageShell>
</template>
<script setup lang="ts">
import { useRoute } from "vue-router";
import AppIcon from "@/shared/components/AppIcon.vue";
import PageShell from "@/shared/components/PageShell.vue";
import { styleLibraryPages } from "../catalog";
import "@/features/movies/styles/index.css";
import "@/features/fc2/styles/index.css";
import "../styles/index.css";
const route = useRoute();
</script>
