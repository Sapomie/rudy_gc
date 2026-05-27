<template>
  <PageFrame title="页面未找到" eyebrow="404" description="当前路径没有 v2 页面。">
    <div class="rgc-panel">
      <RouterLink class="rgc-button primary" to="/cards">回到 Cards</RouterLink>
    </div>
  </PageFrame>
</template>

<script setup lang="ts">
import PageFrame from "@/components/PageFrame.vue";
</script>
