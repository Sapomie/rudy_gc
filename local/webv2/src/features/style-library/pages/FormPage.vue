<template>
  <div class="style-library-view">
    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>按钮与操作状态</h2>
        <p>悬停、按下、键盘焦点、禁用</p>
      </header>
      <div class="style-library-section-body">
        <div class="style-library-actions">
          <button type="button" class="button-primary" @click="notice = '已执行主要操作'">主要操作</button
          ><button type="button" class="button-secondary" @click="notice = '已执行次要操作'">次要操作</button
          ><button type="button" class="button-danger" @click="notice = '危险操作示例：请在反馈页查看二次确认'">
            危险操作</button
          ><button
            type="button"
            class="button-secondary active"
            aria-pressed="true"
            @click="notice = '当前按钮处于选中状态'"
          >
            已选中</button
          ><button type="button" class="button-primary" disabled>不可用</button
          ><button type="button" class="button-secondary" disabled>处理中…</button>
        </div>
        <div class="style-library-columns">
          <div class="style-library-sample route-movie-detail">
            <h3>稍后下载</h3>
            <div class="style-library-actions">
              <button type="button" class="btn movie-download-later-button btn-outline-warning" @click="added = true">
                稍后下载</button
              ><button type="button" class="btn movie-download-later-button btn-success" @click="added = false">
                已添加</button
              ><span class="style-library-note">当前：{{ added ? "已添加" : "未添加" }}</span>
            </div>
          </div>
          <div class="style-library-sample route-sh-magnets">
            <h3>下载相册</h3>
            <div class="fetch-site-album-inline style-library-actions">
              <button
                type="button"
                class="btn"
                :class="downloading ? 'btn-warning' : 'btn-outline-warning'"
                :aria-pressed="downloading"
                @click="downloading = !downloading"
              >
                {{ downloading ? "已在下载中" : "下载中" }}</button
              ><button
                type="button"
                class="btn"
                :class="pending ? 'btn-info' : 'btn-outline-info'"
                :aria-pressed="pending"
                @click="pending = !pending"
              >
                {{ pending ? "已在待下载" : "待下载" }}
              </button>
            </div>
          </div>
        </div>
        <p class="style-library-note" role="status">
          {{ notice || "使用 Tab 查看焦点；点击相册按钮比较有／无状态。" }}
        </p>
      </div>
    </section>
    <div class="style-library-columns">
      <section class="style-library-section">
        <header class="style-library-section-head">
          <h2>表单示例</h2>
          <p>填写、校验、保存与重置</p>
        </header>
        <form class="style-library-section-body style-library-form" @submit.prevent="submit" @reset.prevent="reset">
          <label class="filter-field"
            ><span>名称 <span class="style-library-required">*</span></span
            ><input v-model="name" required placeholder="输入示例名称"
          /></label>
          <label class="filter-field"
            ><span>分类</span
            ><select v-model="category">
              <option>电影相册</option>
              <option>下载相册</option>
              <option>收藏</option>
            </select></label
          >
          <label class="filter-field"><span>日期</span><input v-model="date" type="date" /></label
          ><label class="filter-field"><span>只读编号</span><input value="STYLE-001" readonly /></label>
          <label class="filter-field style-library-form-wide"
            ><span>说明</span><textarea v-model="description" rows="3" placeholder="补充说明" />
          </label>
          <div class="style-library-checks style-library-form-wide">
            <label><input v-model="enabled" type="checkbox" />启用示例</label
            ><label><input v-model="visibility" value="private" type="radio" name="visibility" />仅自己</label
            ><label><input v-model="visibility" value="shared" type="radio" name="visibility" />共享</label>
          </div>
          <div class="style-library-actions style-library-form-wide">
            <button class="button-primary" type="submit">保存示例</button
            ><button class="button-secondary" type="reset">重置</button>
          </div>
          <p v-if="saved" class="state-line success style-library-form-wide" role="status">{{ saved }}</p>
        </form>
      </section>
      <section class="style-library-section">
        <header class="style-library-section-head">
          <h2>输入状态</h2>
          <p>清楚表达可编辑性</p>
        </header>
        <div class="style-library-section-body">
          <label class="filter-field"><span>普通输入</span><input placeholder="点击或使用 Tab 聚焦" /></label>
          <label class="filter-field"><span>禁用输入</span><input value="等待前置任务完成" disabled /></label>
          <label class="filter-field"
            ><span>字段错误</span
            ><input
              value=""
              aria-invalid="true"
              aria-describedby="style-field-error"
              class="style-library-invalid"
              placeholder="必填内容"
            /><small id="style-field-error" class="style-library-required">请输入名称后继续。</small></label
          >
          <div class="style-library-checks">
            <label><input type="checkbox" checked disabled />已选 / 禁用</label
            ><label><input type="checkbox" disabled />未选 / 禁用</label>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from "vue";
const notice = ref("");
const added = ref(false);
const pending = ref(true);
const downloading = ref(false);
const name = ref("");
const category = ref("电影相册");
const date = ref("2026-10-23");
const description = ref("");
const enabled = ref(true);
const visibility = ref("private");
const saved = ref("");
function submit() {
  saved.value = `已保存「${name.value}」· ${category.value} · ${enabled.value ? "启用" : "停用"}`;
}
function reset() {
  name.value = "";
  category.value = "电影相册";
  date.value = "2026-10-23";
  description.value = "";
  enabled.value = true;
  visibility.value = "private";
  saved.value = "";
}
</script>
