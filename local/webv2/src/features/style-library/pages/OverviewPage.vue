<template>
  <div class="style-library-view">
    <div class="style-library-summary">
      <div>
        <span>主题</span><strong><i class="style-library-theme-dot" />玫瑰粉</strong>
      </div>
      <div>
        <span>列表密度</span><strong>30 px <small>行高</small></strong>
      </div>
      <div><span>每行卡片</span><strong>6 / 8 / 10</strong></div>
      <div>
        <span>页面分类</span><strong>6 <small>类</small></strong>
      </div>
    </div>
    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>页面目录</h2>
        <p>浏览组件、比较状态、试用交互</p>
      </header>
      <div class="style-library-page-grid">
        <RouterLink
          v-for="page in styleLibraryPages.filter((item) => item.path !== '/styles')"
          :key="page.path"
          :to="page.path"
          class="style-library-page-card"
        >
          <span class="style-library-page-card-icon"><AppIcon :name="page.icon" /></span>
          <span class="style-library-page-card-copy"
            ><strong>{{ page.title }}</strong
            ><small>{{ page.description }}</small></span
          >
          <span class="style-library-page-card-arrow" aria-hidden="true">›</span>
        </RouterLink>
      </div>
    </section>
    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>常用状态速览</h2>
        <p>颜色与文字共同表达状态</p>
      </header>
      <div class="style-library-section-body style-library-columns">
        <div class="style-library-sample">
          <h3>任务与数据</h3>
          <div class="style-library-actions">
            <StatusBadge v-for="state in semanticStates" :key="state.label" v-bind="state" />
          </div>
        </div>
        <div class="style-library-sample">
          <h3>播放与角标</h3>
          <div class="style-library-actions movie-card-media-block">
            <span class="btn btn-success">Play · 有字幕</span><span class="btn btn-info">Play · 无字幕</span
            ><span class="badge-come">❤ 2</span><span class="badge-times">× 3</span>
          </div>
        </div>
      </div>
    </section>
    <p class="style-library-note">示例使用本地演示数据，操作结果仅在样式库中展示。</p>
  </div>
</template>
<script setup lang="ts">
import AppIcon from "@/shared/components/AppIcon.vue";
import StatusBadge from "@/features/remote-dailybest/components/StatusBadge.vue";
import { semanticStates, styleLibraryPages } from "../catalog";
</script>
