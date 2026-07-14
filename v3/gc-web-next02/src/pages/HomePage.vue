<template>
  <PageShell title="Rudy-GC Next02" note="当前 Vue 前端入口。">
    <div class="home-grid">
      <section v-for="group in sidebarGroups" :key="group.label" class="home-section">
        <h2>{{ group.label }}</h2>
        <div class="home-links">
          <RouterLink v-for="link in group.links" :key="link.href" :to="link.href">
            {{ link.label }}
          </RouterLink>
        </div>
      </section>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import PageShell from "@/components/PageShell.vue";
import { sidebarGroups } from "@/data/navigation";
</script>
