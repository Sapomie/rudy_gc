<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { loadHealthState, type HealthState } from "@/features/dailybest/api";

const health = ref<HealthState>("loading");
const label = computed(() => ({ loading: "检查中", ready: "运行正常", unavailable: "服务不可用" })[health.value]);

onMounted(async () => {
  try {
    health.value = await loadHealthState();
  } catch {
    health.value = "unavailable";
  }
});
</script>

<template>
  <main class="page-shell">
    <section class="hero">
      <p class="eyebrow">REMOTE SERVICE</p>
      <h1>DailyBest 管理端</h1>
      <p class="summary">远端采集、快照发布和同步接口的独立管理入口。</p>
      <div class="status" :data-state="health">
        <span class="status-dot" aria-hidden="true"></span>
        <span>{{ label }}</span>
      </div>
    </section>
  </main>
</template>
