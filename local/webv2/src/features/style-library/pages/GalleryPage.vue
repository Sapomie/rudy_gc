<template>
  <div class="style-library-view cards-page-shell next02-pink-cards-page">
    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>卡片与相框</h2>
        <p>2:3 封面 · 左上角标 · 右下评分</p>
      </header>
      <div class="cards-toolbar card-results-toolbar next02-pink-cards-toolbar">
        <div class="cards-toolbar-sort sort-controls">
          <div class="link-strip">
            <SortLink
              v-for="option in sorts"
              :key="option.key"
              class="strip-link"
              href="#"
              :label="option.label"
              :active="sort === option.key"
              :direction="descending ? 'desc' : 'asc'"
              @click.prevent="setSort(option.key)"
            />
          </div>
        </div>
        <div class="toolbar-line cards-toolbar-bottom">
          <div class="owned-line cards-toolbar-filters">
            <div class="link-strip">
              <button
                v-for="option in kinds"
                :key="option.value"
                type="button"
                class="strip-link"
                :class="{ active: kind === option.value }"
                :aria-pressed="kind === option.value"
                @click="kind = option.value"
              >
                {{ option.label }}
              </button>
            </div>
            <label class="style-library-search"
              ><input v-model="keyword" aria-label="搜索卡片" placeholder="搜索示例名称" type="search"
            /></label>
          </div>
          <div class="next02-pink-cards-toolbar-status">
            <div class="card-grid-size-control">
              <div class="rank-navigation-group" role="group" aria-label="每行卡片数快捷选项">
                <button
                  v-for="count in [6, 8, 10]"
                  :key="count"
                  type="button"
                  class="button-secondary rank-toolbar-action"
                  :class="{ active: columns === count }"
                  :aria-pressed="columns === count"
                  @click="columns = count"
                >
                  {{ count }}
                </button>
              </div>
            </div>
            <div class="cards-toolbar-pagination">
              <span class="style-library-note">{{ cards.length }} 张示例 · 点击查看详情</span>
            </div>
          </div>
        </div>
      </div>
    </section>
    <div
      class="movie-grid-v2 next02-pink-movie-grid"
      :style="{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }"
      @click.capture="previewCard"
    >
      <div
        v-for="item in cards"
        :key="item.id"
        class="col movie-grid-col-v2"
        style="width: 100%; max-width: 100%; flex: none; min-width: 0"
      >
        <MovieCard v-if="item.kind === 'movie'" :movie="item.movie" /><FC2MovieCard v-else :movie="item.movie" />
      </div>
      <div v-if="!cards.length" class="movie-empty" style="grid-column: 1 / -1">
        没有匹配的示例。<button
          type="button"
          class="button-secondary"
          @click="
            keyword = '';
            kind = 'all';
          "
        >
          清除筛选
        </button>
      </div>
    </div>
    <p class="style-library-note">
      示例涵盖有字幕、无字幕、单／双角标、缺失封面和移除状态；卡片入口统一打开样式库详情示例。
    </p>
  </div>
</template>
<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import MovieCard from "@/features/movies/components/MovieCard.vue";
import FC2MovieCard from "@/features/fc2/components/FC2MovieCard.vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import { galleryCards } from "../fixtures";
const router = useRouter();
const columns = ref(6);
const keyword = ref("");
const kind = ref("all");
const sort = ref("id");
const descending = ref(false);
const kinds = [
  { label: "全部", value: "all" },
  { label: "Movie", value: "movie" },
  { label: "FC2", value: "fc2" }
];
const sorts = [
  { label: "默认顺序", key: "id" },
  { label: "名称", key: "name" }
];
const cards = computed(() =>
  galleryCards
    .filter(
      (item) =>
        (kind.value === "all" || item.kind === kind.value) &&
        `${item.name} ${item.title}`.toLowerCase().includes(keyword.value.trim().toLowerCase())
    )
    .sort((a, b) => (sort.value === "name" ? a.name.localeCompare(b.name) : a.id - b.id) * (descending.value ? -1 : 1))
);
function setSort(key: string) {
  descending.value = sort.value === key ? !descending.value : false;
  sort.value = key;
}
function previewCard(event: MouseEvent) {
  if (!(event.target instanceof Element) || !event.target.closest("a")) return;
  event.preventDefault();
  event.stopPropagation();
  void router.push("/styles/detail");
}
</script>
