<template>
  <PageShell class="cards-agg-page-shell next02-pink-cards-page next02-pink-cards-page-shell" :title="title" :note="note" eyebrow="MovieCard Agg">
    <template #actions>
      <a :href="cardsHref" class="next02-pink-cards-hero-link" target="_blank" rel="noopener noreferrer">Cards</a>
    </template>
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <template v-else>
      <div class="sticky-toolbar cards-agg-toolbar next02-pink-cards-toolbar">
        <div class="owned-line cards-agg-owned next02-pink-cards-agg-controls">
          <span class="soft-badge cards-source-badge disabled">WMedia</span>
          <a
            v-for="item in ownedItems"
            :key="item.key"
            :href="item.href || '#'"
            class="btn btn-sm waves-effect waves-light"
            :class="[item.className, { active: item.active, disabled: item.active }]"
          >
            {{ item.label }}
          </a>
          <button
            class="button-secondary cards-filter-button"
            :class="{ active: filterOpen }"
            type="button"
            :aria-pressed="filterOpen"
            data-bs-toggle="collapse"
            data-bs-target="#movieCardFilterCollapse"
            @click="filterOpen = !filterOpen"
          >
            筛选
          </button>
        </div>
        <div class="next02-pink-cards-metrics" aria-label="聚合页面摘要">
          <span><strong>{{ totalLabel }}</strong> Total</span>
        </div>
      </div>

      <MovieCardFilterForm
        :filter="filter"
        form-id="cardsAggFilterForm"
        variant="cards-agg"
        collapse-id="movieCardFilterCollapse"
        :open="filterOpen"
      />

      <section v-if="skipped" class="section-card">
        <div class="alert alert-info mb-0">
          <div class="fw-semibold">{{ message || "当前条件已跳过聚合。" }}</div>
          <div v-if="skipMeta" class="text-muted small mt-2">{{ skipMeta }}</div>
        </div>
      </section>

      <template v-else>
        <section class="mb-3">
          <div class="cards-agg-summary-grid">
            <div v-for="section in summaryTables" :key="section.key" class="col">
              <AggItemTable :title="section.title" :items="section.items" />
            </div>
          </div>
        </section>

        <section class="mb-3">
          <AggCastTable title="演员" :items="listOf(groups, 'Casts')" />
        </section>

        <div class="cards-agg-two-col">
          <div class="col"><AggItemTable title="导演" :items="listOf(groups, 'Directors')" /></div>
          <div class="col"><AggItemTable title="前缀" :items="listOf(groups, 'Prefixes')" /></div>
          <div class="col"><AggItemTable title="片商" :items="listOf(groups, 'Makers')" /></div>
          <div class="col"><AggItemTable title="标签" :items="listOf(groups, 'Labels')" /></div>
        </div>

        <div class="cards-agg-two-col mt-1">
          <div class="col"><AggItemTable title="上映年份" :items="listOf(timeline, 'ReleaseYears')" /></div>
          <div class="col"><AggItemTable title="上映年月" :items="listOf(timeline, 'ReleaseMonths')" /></div>
          <div class="col"><AggItemTable title="Media 下载年份" :items="listOf(timeline, 'MediaBirthYears')" /></div>
          <div class="col"><AggItemTable title="Media 下载年月" :items="listOf(timeline, 'MediaBirthMonths')" /></div>
          <div class="col"><AggItemTable title="最后 SC 年月" :items="listOf(timeline, 'LastScMonths')" /></div>
        </div>

        <div class="cards-agg-two-col mt-1">
          <div class="col"><AggItemTable title="评分分布" :items="listOf(distributions, 'Scores')" /></div>
          <div class="col"><AggItemTable title="观看人数分布" :items="listOf(distributions, 'Viewers')" /></div>
          <div class="col"><AggItemTable title="片长分布" :items="listOf(distributions, 'Lengths')" /></div>
          <div class="col"><AggItemTable title="SC 次数分布" :items="listOf(distributions, 'ScTimes')" /></div>
          <div class="col"><AggItemTable title="Come 次数分布" :items="listOf(distributions, 'ComeTimes')" /></div>
          <div class="col"><AggItemTable title="最高排名分布" :items="listOf(distributions, 'HighestRanks')" /></div>
        </div>

        <div class="cards-agg-two-col mt-1">
          <div class="col"><AggTopMovieTable title="SC 次数" :items="listOf(topMovies, 'ScTimes')" /></div>
          <div class="col"><AggTopMovieTable title="Come 次数" :items="listOf(topMovies, 'ComeTimes')" /></div>
          <div class="col"><AggTopMovieTable title="观看人数" :items="listOf(topMovies, 'Viewers')" /></div>
          <div class="col"><AggTopMovieTable title="评分" :items="listOf(topMovies, 'Scores')" /></div>
          <div class="col"><AggTopMovieTable title="最高排名" :items="listOf(topMovies, 'HighestRanks')" /></div>
          <div class="col"><AggTopMovieTable title="最近上映" :items="listOf(topMovies, 'RecentRelease')" /></div>
          <div class="col"><AggTopMovieTable title="最近 Media 下载" :items="listOf(topMovies, 'RecentMedia')" /></div>
          <div class="col"><AggTopMovieTable title="最近 SC" :items="listOf(topMovies, 'RecentSc')" /></div>
        </div>
      </template>
    </template>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/shared/api/request";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import MovieCardFilterForm from "@/features/movies/components/MovieCardFilterForm.vue";
import PageShell from "@/shared/components/PageShell.vue";
import type { AnyRecord } from "@/shared/types/gc";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const filterOpen = ref(false);

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  filterOpen.value = false;
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title") || "MovieCard Agg");
const totalLabel = computed(() => text(payload.value, "TotalLabel", "total", "Total") || "0");
const note = computed(() => `Total ${totalLabel.value}`);
const message = computed(() => text(payload.value, "Message", "message"));
const filter = computed(() => asRecord(pick(payload.value, "MovieCardFilter")));
const skipped = computed(() => Boolean(payload.value.Skipped || payload.value.skipped));
const skipMeta = computed(() => {
  const total = text(payload.value, "Total", "total");
  const limit = text(payload.value, "Limit", "limit");
  return text(payload.value, "SkipReason") === "too_many_movies" && total
    ? `当前命中 ${total} 部，上限 ${limit} 部。`
    : "";
});
const cardsHref = computed(() => `/listcard${route.fullPath.includes("?") ? route.fullPath.slice(route.fullPath.indexOf("?")) : ""}`);
const ownedQuery = computed(() => asRecord(pick(payload.value, "ownedQuery", "OwnedQuery")));
const summary = computed(() => asRecord(pick(payload.value, "Summary")));
const groups = computed(() => asRecord(pick(payload.value, "Groups")));
const timeline = computed(() => asRecord(pick(payload.value, "Timeline")));
const distributions = computed(() => asRecord(pick(payload.value, "Distributions")));
const topMovies = computed(() => asRecord(pick(payload.value, "TopMovies")));

const ownedItems = computed(() => [
  ownedItem("MediaAll", "MAll", "btn-primary"),
  ownedItem("MediaOwned", "MOwned", "btn-success"),
  ownedItem("MediaNotOwned", "MNotOwned", "btn-secondary")
]);

const summaryTables = computed(() => [
  { key: "Owned", title: "WMedia", items: listOf(summary.value, "Owned") },
  { key: "Sc", title: "SC", items: listOf(summary.value, "Sc") },
  { key: "Come", title: "Come", items: listOf(summary.value, "Come") },
  { key: "Rank", title: "榜单", items: listOf(summary.value, "Rank") }
]);

function ownedItem(key: string, label: string, className: string) {
  const item = asRecord(ownedQuery.value[key]);
  return {
    key,
    label,
    className,
    href: text(item, "Href", "href") || "#",
    active: Boolean(item.Active || item.active)
  };
}

function listOf(obj: AnyRecord, key: string): AnyRecord[] {
  return asArray<AnyRecord>(obj[key] ?? obj[key.charAt(0).toLowerCase() + key.slice(1)]);
}

function itemLabel(item: AnyRecord): string {
  return text(item, "Label", "label", "Key", "key");
}

function itemCount(item: AnyRecord): string {
  return text(item, "Count", "count") || "0";
}

function itemPercent(item: AnyRecord): string {
  const percent = text(item, "Percent", "percent");
  return `${percent}%`;
}

function itemHref(item: AnyRecord): string {
  return text(item, "Href", "href");
}

function topMovieHref(item: AnyRecord): string {
  return text(item, "Href", "href") || `/moviedetail/${encodeURIComponent(text(item, "Name", "JavID"))}`;
}

const AggItemTable = defineComponent({
  props: {
    title: { type: String, required: true },
    items: { type: Array as PropType<AnyRecord[]>, default: () => [] }
  },
  setup(props) {
    return () => h("section", { class: "section-card next02-pink-cards-agg-panel" }, [
      h("h4", { class: "section-title" }, props.title),
      h("div", { class: "table-responsive" }, [
        h("table", { class: "table table-sm table-hover align-middle mb-0 next02-pink-cards-agg-table agg-item-table" }, [
          h("thead", [
            h("tr", [
              h("th", "名称"),
              h("th", { class: "text-end" }, "数量"),
              h("th", { class: "text-end" }, "占比")
            ])
          ]),
          h("tbody", props.items.length ? props.items.map((row) => h("tr", [
            h("td", { class: "text-nowrap" }, itemHref(row)
              ? h("a", { href: itemHref(row), target: "_blank", rel: "noopener noreferrer" }, itemLabel(row))
              : itemLabel(row)),
            h("td", { class: "text-end" }, itemCount(row)),
            h("td", { class: "text-end" }, itemPercent(row))
          ])) : [
            h("tr", [h("td", { colspan: 3, class: "text-muted" }, "暂无数据")])
          ])
        ])
      ])
    ]);
  }
});

const AggCastTable = defineComponent({
  props: {
    title: { type: String, required: true },
    items: { type: Array as PropType<AnyRecord[]>, default: () => [] }
  },
  setup(props) {
    return () => h("section", { class: "section-card next02-pink-cards-agg-panel" }, [
      h("h4", { class: "section-title" }, props.title),
      h("div", { class: "table-responsive" }, [
        h("table", { class: "table table-sm table-hover align-middle mb-0 next02-pink-cards-agg-table agg-cast-table" }, [
          h("thead", [
            h("tr", [
              h("th", "名称"),
              h("th", { class: "text-end" }, "总数"),
              h("th", { class: "text-end" }, "占比"),
              h("th", { class: "text-end" }, "有 SC"),
              h("th", { class: "text-end" }, "无 SC"),
              h("th", { class: "text-end" }, "MOwned"),
              h("th", { class: "text-end" }, "MRemoved"),
              h("th", { class: "text-end" }, "MNotOwned")
            ])
          ]),
          h("tbody", props.items.length ? props.items.map((row) => h("tr", [
            h("td", { class: "text-nowrap" }, itemHref(row)
              ? h("a", { href: itemHref(row), target: "_blank", rel: "noopener noreferrer" }, itemLabel(row))
              : itemLabel(row)),
            h("td", { class: "text-end" }, itemCount(row)),
            h("td", { class: "text-end" }, itemPercent(row)),
            h("td", { class: "text-end" }, text(row, "ScCount") || "0"),
            h("td", { class: "text-end" }, text(row, "NoScCount") || "0"),
            h("td", { class: "text-end" }, text(row, "MOwnedCount") || "0"),
            h("td", { class: "text-end" }, text(row, "MRemovedCount") || "0"),
            h("td", { class: "text-end" }, text(row, "MNotOwnedCount") || "0")
          ])) : [
            h("tr", [h("td", { colspan: 8, class: "text-muted" }, "暂无数据")])
          ])
        ])
      ])
    ]);
  }
});

const AggTopMovieTable = defineComponent({
  props: {
    title: { type: String, required: true },
    items: { type: Array as PropType<AnyRecord[]>, default: () => [] }
  },
  setup(props) {
    return () => h("section", { class: "section-card next02-pink-cards-agg-panel" }, [
      h("h4", { class: "section-title" }, props.title),
      h("div", { class: "table-responsive" }, [
        h("table", { class: "table table-sm table-hover align-middle mb-0 next02-pink-cards-agg-table agg-top-movie-table" }, [
          h("thead", [
            h("tr", [
              h("th", "番号"),
              h("th", "标题"),
              h("th", { class: "text-end" }, "值"),
              h("th", "日期")
            ])
          ]),
          h("tbody", props.items.length ? props.items.map((row) => h("tr", [
            h("td", { class: "text-nowrap" }, h("a", { href: topMovieHref(row), target: "_blank", rel: "noopener noreferrer" }, text(row, "Name", "JavID"))),
            h("td", text(row, "Title")),
            h("td", { class: "text-end text-nowrap" }, text(row, "ValueLabel", "Value")),
            h("td", { class: "text-nowrap" }, text(row, "Date"))
          ])) : [
            h("tr", [h("td", { colspan: 4, class: "text-muted" }, "暂无数据")])
          ])
        ])
      ])
    ]);
  }
});
</script>
