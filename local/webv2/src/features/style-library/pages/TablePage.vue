<template>
  <div class="style-library-view">
    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>紧凑数据列表</h2>
        <p>30px 行高 · 悬停与选中 · 本页多选</p>
      </header>
      <div class="style-library-toolbar">
        <label class="style-library-field"
          ><span>名称</span><input v-model="keyword" type="search" placeholder="搜索示例"
        /></label>
        <label class="style-library-field"
          ><span>状态</span
          ><select v-model="status">
            <option value="all">全部状态</option>
            <option v-for="(state, index) in semanticStates" :key="state.label" :value="String(index)">
              {{ state.label }}
            </option>
          </select></label
        >
        <button
          type="button"
          class="button-secondary"
          @click="
            keyword = '';
            status = 'all';
          "
        >
          重置
        </button>
        <span class="style-library-toolbar-spacer" />
        <span class="style-library-note" role="status">已选 {{ selectedCount }} 项</span>
        <button type="button" class="button-secondary" :disabled="!selectedCount" @click="completeSelection">
          标记已完成
        </button>
        <button type="button" class="button-secondary" :disabled="!selectedCount" @click="clear">清除选择</button>
      </div>
      <div class="table-scroll style-library-table-wrap">
        <table class="data-table table style-library-data-table">
          <thead>
            <tr>
              <th class="style-library-check-cell">
                <input
                  type="checkbox"
                  aria-label="选择本页全部"
                  :checked="allSelected"
                  :indeterminate="partialSelected"
                  @change="toggleAll(($event.target as HTMLInputElement).checked)"
                />
              </th>
              <th :aria-sort="sortKey === 'name' ? direction : 'none'">
                <SortLink
                  href="#"
                  label="名称"
                  :active="sortKey === 'name'"
                  :direction="descending ? 'desc' : 'asc'"
                  @click.prevent="setSort('name')"
                />
              </th>
              <th>任务</th>
              <th :aria-sort="sortKey === 'date' ? direction : 'none'">
                <SortLink
                  href="#"
                  label="日期"
                  :active="sortKey === 'date'"
                  :direction="descending ? 'desc' : 'asc'"
                  @click.prevent="setSort('date')"
                />
              </th>
              <th class="sc-number" :aria-sort="sortKey === 'count' ? direction : 'none'">
                <SortLink
                  href="#"
                  label="数量"
                  :active="sortKey === 'count'"
                  :direction="descending ? 'desc' : 'asc'"
                  @click.prevent="setSort('count')"
                />
              </th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in visibleRows" :key="row.id" :class="{ 'is-selected': isSelected(row) }">
              <td class="style-library-check-cell">
                <input
                  type="checkbox"
                  :aria-label="`选择 ${row.name}`"
                  :checked="isSelected(row)"
                  @change="toggleRow(row, ($event.target as HTMLInputElement).checked)"
                />
              </td>
              <td>
                <RouterLink to="/styles/detail">{{ row.name }}</RouterLink>
              </td>
              <td>{{ row.title }}</td>
              <td><ScDateTime :value="row.date" /></td>
              <td class="sc-number">{{ row.count }}</td>
              <td><StatusBadge v-bind="semanticStates[row.status]" /></td>
              <td>
                <button
                  type="button"
                  class="button-secondary table-action"
                  @click="message = `${row.name}：${semanticStates[row.status].label}`"
                >
                  查看
                </button>
              </td>
            </tr>
            <tr v-if="!visibleRows.length">
              <td colspan="7"><div class="state-line">没有匹配的示例，请调整筛选条件。</div></td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="style-library-toolbar">
        <span class="style-library-note">共 {{ filteredRows.length }} 条 · 第 {{ page }} / {{ pageCount }} 页</span
        ><label class="style-library-field"
          ><span>每页</span
          ><select v-model.number="pageSize">
            <option :value="6">6</option>
            <option :value="12">12</option>
            <option :value="18">18</option>
          </select></label
        ><span class="style-library-toolbar-spacer" /><PaginationBar :links="pageLinks" @click="changePage" />
      </div>
    </section>
    <p class="style-library-note" role="status">{{ message || "点击表头排序；筛选或翻页后清除本页选择。" }}</p>
  </div>
</template>
<script setup lang="ts">
import { computed, ref, watch } from "vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import StatusBadge from "@/features/remote-dailybest/components/StatusBadge.vue";
import { useAlbumSelection } from "@/features/movies/composables/useAlbumSelection";
import { semanticStates } from "../catalog";
import { tableRows } from "../fixtures";
const rows = ref(tableRows.map((row) => ({ ...row })));
const keyword = ref("");
const status = ref("all");
const page = ref(1);
const pageSize = ref(6);
const message = ref("");
const sortKey = ref<"name" | "date" | "count">("date");
const descending = ref(true);
const direction = computed(() => (descending.value ? "descending" : "ascending"));
const filteredRows = computed(() =>
  rows.value
    .filter(
      (row) =>
        (status.value === "all" || String(row.status) === status.value) &&
        `${row.name} ${row.title}`.toLowerCase().includes(keyword.value.trim().toLowerCase())
    )
    .sort((a, b) => {
      const left = a[sortKey.value];
      const right = b[sortKey.value];
      return (
        (typeof left === "number" && typeof right === "number"
          ? left - right
          : String(left).localeCompare(String(right))) * (descending.value ? -1 : 1)
      );
    })
);
const pageCount = computed(() => Math.max(1, Math.ceil(filteredRows.value.length / pageSize.value)));
const visibleRows = computed(() =>
  filteredRows.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value)
);
const { selectedCount, selectedIDs, allSelected, partialSelected, isSelected, toggleRow, toggleAll, clear } =
  useAlbumSelection(visibleRows, (row) => row.id);
const pageLinks = computed(() => [
  { Label: "上一页", Href: `?p=${page.value - 1}`, Disabled: page.value === 1 },
  ...Array.from({ length: pageCount.value }, (_, index) => ({
    Label: String(index + 1),
    Href: `?p=${index + 1}`,
    Active: page.value === index + 1
  })),
  { Label: "下一页", Href: `?p=${page.value + 1}`, Disabled: page.value === pageCount.value }
]);
watch([keyword, status, pageSize], () => {
  page.value = 1;
  clear();
});
watch([page, sortKey, descending], clear);
function setSort(key: typeof sortKey.value) {
  descending.value = sortKey.value === key ? !descending.value : false;
  sortKey.value = key;
}
function changePage(event: MouseEvent) {
  const anchor = (event.target as Element).closest("a");
  if (!anchor) return;
  event.preventDefault();
  page.value = Number(new URL(anchor.href).searchParams.get("p") || 1);
}
function completeSelection() {
  const count = selectedCount.value;
  rows.value.forEach((row) => {
    if (selectedIDs.value.has(row.id)) row.status = 0;
  });
  clear();
  page.value = Math.min(page.value, pageCount.value);
  message.value = `已将 ${count} 条示例标记为已完成`;
}
</script>
