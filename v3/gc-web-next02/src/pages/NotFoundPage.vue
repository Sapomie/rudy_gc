<template>
  <PageShell title="页面不存在" :note="route.fullPath" eyebrow="404">
    <div class="empty-block">当前路径没有匹配到 next02 路由。</div>
  </PageShell>
</template>

<script setup lang="ts">
import { useRoute } from "vue-router";
import PageShell from "@/components/PageShell.vue";

const route = useRoute();
</script>
