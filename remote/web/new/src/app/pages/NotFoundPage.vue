<template>
  <PageShell eyebrow="REMOTE" title="页面不存在" note="当前地址没有对应的管理页面。">
    <section class="panel empty-panel">
      <p>返回运行概览继续操作。</p>
      <RouterLink class="button-primary" to="/dailybestoverview">返回概览</RouterLink>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import PageShell from "@/shared/components/PageShell.vue";
</script>
