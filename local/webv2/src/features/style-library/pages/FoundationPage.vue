<template>
  <div class="style-library-view">
    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>颜色</h2>
        <p>主题色、语义色与界面层次</p>
      </header>
      <div class="style-library-section-body">
        <div class="style-library-color-grid">
          <div v-for="color in colors" :key="color.token" class="style-library-color">
            <span class="style-library-color-swatch" :style="{ background: `var(${color.token})` }" />
            <div>
              <strong>{{ color.label }}</strong
              ><code>{{ color.token }}</code>
            </div>
          </div>
        </div>
        <div class="style-library-actions">
          <StatusBadge v-for="state in semanticStates" :key="state.label" v-bind="state" />
        </div>
      </div>
    </section>
    <div class="style-library-columns">
      <section class="style-library-section">
        <header class="style-library-section-head">
          <h2>文字层级</h2>
          <p>紧凑、清晰</p>
        </header>
        <div class="style-library-section-body style-library-type-scale">
          <div><code>16 px</code><strong style="font-size: 16px">页面与面板标题</strong></div>
          <div><code>14 px</code><strong style="font-size: 14px">影片名称与重要数值</strong></div>
          <div><code>12 px</code><span>列表内容、字段与按钮</span></div>
          <div><code>11 px</code><small class="style-library-note">日期、说明与辅助信息</small></div>
        </div>
      </section>
      <section class="style-library-section">
        <header class="style-library-section-head">
          <h2>边界与间距</h2>
          <p>沿用页面公共尺寸</p>
        </header>
        <div class="style-library-section-body style-library-actions">
          <div class="style-library-dimension" style="border-radius: var(--control-radius)">
            <strong>5 px</strong><small>控件圆角</small>
          </div>
          <div class="style-library-dimension"><strong>7 px</strong><small>卡片圆角</small></div>
          <div class="style-library-dimension"><strong>1 px</strong><small>分隔边框</small></div>
          <div class="style-library-dimension"><strong>6 / 12 / 16</strong><small>控件 / 区块 / 面板间距</small></div>
        </div>
      </section>
    </div>
    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>项目图标</h2>
        <p>导航与操作使用同一套线性图标</p>
      </header>
      <div class="style-library-section-body style-library-icon-grid">
        <div v-for="icon in icons" :key="icon.name" class="style-library-icon">
          <AppIcon :name="icon.name" /><span>{{ icon.label }}</span>
        </div>
      </div>
    </section>
  </div>
</template>
<script setup lang="ts">
import AppIcon from "@/shared/components/AppIcon.vue";
import StatusBadge from "@/features/remote-dailybest/components/StatusBadge.vue";
import { semanticStates } from "../catalog";
const colors = [
  { label: "主色", token: "--primary" },
  { label: "主色悬停", token: "--primary-hover" },
  { label: "主色浅底", token: "--primary-soft" },
  { label: "成功 / 有字幕", token: "--success" },
  { label: "信息 / 无字幕", token: "--play-no-sub" },
  { label: "警告", token: "--warn" },
  { label: "错误 / 危险", token: "--danger" },
  { label: "页面背景", token: "--bg" },
  { label: "面板", token: "--panel" },
  { label: "分隔线", token: "--line" },
  { label: "正文", token: "--text" },
  { label: "辅助文字", token: "--muted" }
];
const icons = [
  { name: "photos", label: "图库" },
  { name: "albums", label: "相册" },
  { name: "chart", label: "统计" },
  { name: "tasks", label: "任务" },
  { name: "database", label: "数据" },
  { name: "capture", label: "抓取" },
  { name: "time", label: "时间" },
  { name: "download", label: "下载" },
  { name: "archive", label: "归档" },
  { name: "trash", label: "删除" },
  { name: "text", label: "详情" },
  { name: "palette", label: "样式" }
];
</script>
