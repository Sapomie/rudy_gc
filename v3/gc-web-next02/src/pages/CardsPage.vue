<template>
  <PageShell :class="['cards-page-shell', 'next02-pink-cards-page', 'next02-pink-cards-page-shell', pageClass]" :title="title" :note="summaryText" eyebrow="MovieCard">
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <template v-else>
      <div class="sticky-toolbar cards-toolbar next02-pink-cards-toolbar" aria-label="Cards 控制栏">
        <div class="cards-toolbar-sort">
          <LinkStrip :items="sortLinks" />
        </div>
        <div class="toolbar-line cards-toolbar-bottom">
          <div class="owned-line cards-toolbar-filters">
            <LinkStrip :items="ownedLinks" />
            <button class="button-secondary cards-filter-button" :class="{ active: filterOpen }" type="button" :aria-pressed="filterOpen" @click="filterOpen = !filterOpen">
              筛选
            </button>
            <button
              v-if="hasPageDateEdit"
              class="button-secondary page-date-button cards-date-button"
              type="button"
              @click="openPageDateModal"
            >
              {{ pageDateLabel }}：{{ pageDateDisplay || "未设置" }}
            </button>
          </div>
          <div class="next02-pink-cards-toolbar-status">
            <div class="next02-pink-cards-metrics" aria-label="Cards 页面摘要">
              <span><strong>{{ total }}</strong> Total</span>
              <span><strong>{{ movies.length }}</strong> Current Page</span>
            </div>
            <div class="cards-toolbar-pagination next02-pink-cards-pagination">
              <PaginationBar :page-info="pageInfo" />
            </div>
          </div>
        </div>
      </div>

      <div v-show="filterOpen" class="next02-pink-cards-filter">
        <MovieCardFilterForm :filter="filter" />
      </div>

      <MovieGrid :cards="movies" />

      <div
        v-if="hasPageDateEdit"
        v-show="dateModalOpen"
        class="modal-backdrop-v2"
        role="presentation"
        @click.self="closePageDateModal"
      >
        <form class="modal-panel-v2" @submit.prevent="savePageDate">
          <div class="modal-header-v2">
            <h2>修改{{ pageDateLabel }}</h2>
            <button class="modal-close-v2" type="button" aria-label="关闭" @click="closePageDateModal">×</button>
          </div>
          <div class="modal-body-v2">
            <label class="filter-field">
              <span>日期</span>
              <input
                v-model="pageDateInput"
                id="movieCardPageDateInput"
                type="text"
                placeholder="YYYY-MM-DD"
                inputmode="numeric"
                maxlength="10"
                pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                @input="normalizePageDateInput"
              >
            </label>
            <button class="button-secondary" type="button" @click="setTodayDate">今天</button>
            <div
              v-if="pageDateMessage"
              class="modal-message-v2"
              :class="{ ok: pageDateMessageOK }"
              role="alert"
            >
              {{ pageDateMessage }}
            </div>
          </div>
          <div class="modal-footer-v2">
            <input type="hidden" id="movieCardPageDateKey" :value="text(pageDateEdit, 'ItemKey')">
            <button class="button-secondary" type="button" @click="closePageDateModal">取消</button>
            <button class="button-primary" type="submit" :disabled="pageDateSaving">
              {{ pageDateSaving ? "保存中..." : "保存" }}
            </button>
          </div>
        </form>
      </div>
    </template>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LinkStrip from "@/components/LinkStrip.vue";
import LoadingState from "@/components/LoadingState.vue";
import MovieCardFilterForm from "@/components/MovieCardFilterForm.vue";
import MovieGrid from "@/components/MovieGrid.vue";
import PageShell from "@/components/PageShell.vue";
import PaginationBar from "@/components/PaginationBar.vue";
import type { AnyRecord, PageInfo } from "@/types/gc";
import { asArray, asRecord, pick, text } from "@/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const filterOpen = ref(false);
const dateModalOpen = ref(false);
const pageDateInput = ref("");
const pageDateMessage = ref("");
const pageDateMessageOK = ref(false);
const pageDateSaving = ref(false);
const savedPageDateValue = ref("");

const sortLabel: Record<string, string> = {
  ByDetailUpdate: "Detail更新",
  ByReleasing: "上映日期",
  ByViewer: "观看人数",
  ByRankDate: "上榜日期",
  ByDaysInRank: "在榜天数",
  ByHighestRank: "最高排名",
  ByBirthTime: "下载时间",
  ByLastScTime: "最后sc",
  ByScTimes: "sc次数",
  ByComeTimes: "co次数"
};

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
    dateModalOpen.value = false;
    pageDateMessage.value = "";
    savedPageDateValue.value = "";
  }
}

const cardsPageTitles: Record<string, string> = {
  "/cards": "Cards",
  "/cardstoday": "Cards Today",
  "/cardsmediamowned": "Cards MediaOwned",
  "/cardshasrank": "Cards HasRank",
  "/cardsneeddownload": "Cards NeedDownload",
  "/cardsrandom": "Cards Random",
};
const title = computed(() => cardsPageTitles[route.path] || text(payload.value, "Title") || String(route.meta.title || "MovieCard"));
const pageClass = computed(() => `cards-route-${route.path.replace(/^\/+/, "").replaceAll("/", "-") || "home"}`);
const total = computed(() => text(payload.value, "total", "Total") || "0");
const movies = computed(() => asArray<AnyRecord>(pick(payload.value, "movies", "Movies")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const filter = computed(() => asRecord(pick(payload.value, "MovieCardFilter")));
const pageDateEdit = computed(() => asRecord(pick(payload.value, "PageDateEdit")));
const hasPageDateEdit = computed(() => Boolean(text(pageDateEdit.value, "ItemKey")));
const pageDateLabel = computed(() => text(pageDateEdit.value, "Label") || "日期");
const pageDateStoredValue = computed(() => text(pageDateEdit.value, "Value"));
const pageDateDisplay = computed(() => savedPageDateValue.value || text(pageDateEdit.value, "DisplayValue", "Value"));
const summaryText = computed(() => {
  const display = pageDateDisplay.value;
  return display ? `当前日期：${display}` : "卡片、筛选、排序、分页与旧前端入口保持一致。";
});

const sortLinks = computed(() => {
  const query = asRecord(pick(payload.value, "sortQuery", "SortQuery"));
  return Object.entries(sortLabel).map(([key, label]) => {
    const item = asRecord(query[key]);
    return {
      label,
      href: text(item, "Href") || route.fullPath,
      active: Boolean(item.Active)
    };
  });
});

const ownedLinks = computed(() => {
  const query = asRecord(pick(payload.value, "ownedQuery", "OwnedQuery"));
  const labels: Record<string, string> = {
    MediaAll: "All",
    MediaOwned: "Owned",
    MediaNotOwned: "NotOwned"
  };
  return Object.entries(labels).map(([key, label]) => {
    const item = asRecord(query[key]);
    return {
      label,
      href: text(item, "Href") || route.fullPath,
      active: Boolean(item.Active)
    };
  });
});

function normalizeDateInputValue(raw: string): string {
  const digits = String(raw || "").replace(/\D/g, "").slice(0, 8);
  if (digits.length <= 4) {
    return digits;
  }
  if (digits.length <= 6) {
    return `${digits.slice(0, 4)}-${digits.slice(4)}`;
  }
  return `${digits.slice(0, 4)}-${digits.slice(4, 6)}-${digits.slice(6)}`;
}

function isValidDateValue(raw: string): boolean {
  return /^[0-9]{4}-[0-9]{2}-[0-9]{2}$/.test(String(raw || "").trim());
}

function todayDateString(): string {
  const now = new Date();
  const year = String(now.getFullYear());
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function openPageDateModal() {
  if (!hasPageDateEdit.value) {
    return;
  }
  pageDateInput.value = pageDateStoredValue.value || pageDateDisplay.value;
  pageDateMessage.value = "";
  pageDateMessageOK.value = false;
  dateModalOpen.value = true;
}

function closePageDateModal() {
  if (pageDateSaving.value) {
    return;
  }
  dateModalOpen.value = false;
  pageDateMessage.value = "";
}

function normalizePageDateInput() {
  pageDateInput.value = normalizeDateInputValue(pageDateInput.value);
  pageDateMessage.value = "";
}

function setTodayDate() {
  pageDateInput.value = todayDateString();
  pageDateMessage.value = "";
}

async function savePageDate() {
  const itemKey = text(pageDateEdit.value, "ItemKey");
  const itemValue = pageDateInput.value.trim();
  if (!itemKey) {
    pageDateMessage.value = "缺少日期 key";
    pageDateMessageOK.value = false;
    return;
  }
  if (!itemValue) {
    pageDateMessage.value = "请选择日期";
    pageDateMessageOK.value = false;
    return;
  }
  if (!isValidDateValue(itemValue)) {
    pageDateMessage.value = "日期格式必须为 YYYY-MM-DD";
    pageDateMessageOK.value = false;
    return;
  }
  pageDateSaving.value = true;
  try {
    const result = asRecord(await sendJSON("/api/w-kv/date", "POST", {
      item_key: itemKey,
      item_value: itemValue
    }));
    savedPageDateValue.value = text(result, "item_value") || itemValue;
    pageDateMessage.value = "保存成功";
    pageDateMessageOK.value = true;
    dateModalOpen.value = false;
  } catch (err) {
    pageDateMessage.value = err instanceof Error ? err.message : "保存失败";
    pageDateMessageOK.value = false;
  } finally {
    pageDateSaving.value = false;
  }
}
</script>
