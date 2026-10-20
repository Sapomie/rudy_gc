<template>
  <PageShell
    class="records-page-shell sc-events-page-shell sc-options-page-shell"
    title="SC Options"
    note=""
    eyebrow=""
  >
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <div v-else id="page-root" class="records-page-root next02-sc-pink-theme next02-sc-events-rewrite">
      <div id="scOptionsPage" class="crawl-record-list-page next02-records-rewrite">
        <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">SC / OPTION MANAGEMENT</span>
            <h4>SC Options</h4>
            <p>维护 Kind、FG 和 Vessel 单选项；停用项继续保留历史事件关联。</p>
          </div>
        </section>

        <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
          <div class="card-body">
            <div
              class="rank-stats-toolbar-scope rank-page-shell cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell standard-filter-toolbar crawl-record-filter-toolbar"
            >
              <form
                class="sticky-toolbar cards-toolbar rank-sticky-toolbar next02-pink-cards-toolbar rank-stats-toolbar standard-filter-toolbar-form crawl-record-filter-form"
                method="get"
                action="/scoptionlist"
              >
                <div class="cards-toolbar-sort rank-toolbar-sort standard-filter-toolbar-fields">
                  <div class="rank-toolbar-group standard-filter-group crawl-record-filter-types" aria-label="选项类型">
                    <span class="soft-badge rank-meta-badge">Type</span>
                    <LinkStrip :items="typeLinks" />
                  </div>

                  <div
                    class="rank-toolbar-group standard-filter-group crawl-record-filter-options"
                    aria-label="选项状态"
                  >
                    <span class="soft-badge rank-meta-badge">Status</span>
                    <LinkStrip :items="statusLinks" />
                  </div>
                </div>

                <div class="rank-navigation-group standard-filter-actions crawl-record-filter-actions">
                  <button type="button" class="button-primary rank-toolbar-action" @click="openCreate">新增选项</button>
                </div>

                <div class="toolbar-line cards-toolbar-bottom rank-toolbar-line standard-filter-status-line">
                  <div class="owned-line cards-toolbar-filters rank-toolbar-filters">
                    <span v-if="pageMessage" class="next02-sc-detail-message" role="status">{{ pageMessage }}</span>
                  </div>
                  <div class="next02-pink-cards-toolbar-status">
                    <div class="next02-pink-cards-metrics" aria-label="SC 选项摘要">
                      <span :title="`选项总数 ${items.length}`" :aria-label="`选项总数 ${items.length}`"
                        ><strong>{{ items.length }}</strong></span
                      >
                      <span :title="`启用数量 ${enabledCount}`" :aria-label="`启用数量 ${enabledCount}`"
                        ><strong>{{ enabledCount }}</strong></span
                      >
                      <span
                        :title="`当前展示数量 ${visibleItems.length}`"
                        :aria-label="`当前展示数量 ${visibleItems.length}`"
                        ><strong>{{ visibleItems.length }}</strong></span
                      >
                    </div>
                  </div>
                </div>
              </form>
            </div>

            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 sc-events-table crawl-record-table sc-options-table">
                <colgroup>
                  <col class="sc-options-col-type" />
                  <col class="sc-options-col-name" />
                  <col class="sc-options-col-events" />
                  <col class="sc-options-col-last-event" />
                  <col class="sc-options-col-sort" />
                  <col class="sc-options-col-status" />
                  <col class="sc-options-col-updated" />
                  <col class="sc-options-col-actions" />
                </colgroup>
                <thead>
                  <tr>
                    <th>Type</th>
                    <th>Name</th>
                    <th>SCEventTimes</th>
                    <th>LastSCEventTime</th>
                    <th>SortOrder</th>
                    <th>Status</th>
                    <th>UpdatedOn</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!visibleItems.length">
                    <td colspan="8">
                      <div class="sc-empty-state crawl-record-empty-state">当前筛选下没有 SC Option 数据。</div>
                    </td>
                  </tr>
                  <tr
                    v-for="item in visibleItems"
                    v-else
                    :key="`${item.optionType}:${item.id}`"
                    class="next02-sc-event-row"
                  >
                    <td>
                      <span class="crawl-record-type">{{ optionTypeLabel(item.optionType) }}</span>
                    </td>
                    <td class="sc-content-cell">
                      <a
                        class="sc-event-link crawl-record-name"
                        :href="optionEventsHref(item)"
                        target="_blank"
                        rel="noopener noreferrer"
                        :title="item.optionName"
                      >
                        {{ item.optionName }}
                      </a>
                    </td>
                    <td class="sc-number">{{ item.scEventTimes }}</td>
                    <td :title="formatUnixMinute(item.lastScEventTime)">
                      {{ formatUnixMinute(item.lastScEventTime) }}
                    </td>
                    <td class="sc-number">{{ item.sortOrder }}</td>
                    <td>
                      <span class="sc-option-status" :class="{ 'is-enabled': item.isEnabled === 1 }">
                        {{ item.isEnabled === 1 ? "启用" : "停用" }}
                      </span>
                    </td>
                    <td :title="formatUnixMinute(item.updatedOn)">{{ formatUnixMinute(item.updatedOn) }}</td>
                    <td>
                      <div class="sc-option-actions">
                        <button type="button" class="sc-event-link" @click="openEdit(item)">编辑</button>
                        <button
                          type="button"
                          class="sc-event-link sc-option-disable-link"
                          :disabled="item.isEnabled !== 1"
                          @click="openDelete(item)"
                        >
                          {{ item.isEnabled === 1 ? "停用" : "已停用" }}
                        </button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>

    <Teleport to="body">
      <div
        v-if="editModalOpen"
        class="modal-backdrop-v2 sc-event-edit-overlay"
        role="presentation"
        @click.self="closeEdit"
      >
        <form
          id="scOptionEditForm"
          class="modal-panel-v2 sc-event-edit-panel sc-option-edit-panel"
          role="dialog"
          aria-modal="true"
          aria-labelledby="scOptionEditModalTitle"
          @submit.prevent="saveOption"
        >
          <div class="modal-header-v2">
            <h2 id="scOptionEditModalTitle">{{ editingTarget ? "编辑 SC Option" : "新增 SC Option" }}</h2>
            <button type="button" class="modal-close-v2" aria-label="关闭" :disabled="saving" @click="closeEdit">
              ×
            </button>
          </div>
          <div class="modal-body-v2 sc-event-edit-body">
            <div
              v-if="editMessage"
              class="alert py-2 mb-3"
              :class="editMessageOK ? 'alert-success' : 'alert-danger'"
              role="alert"
            >
              {{ editMessage }}
            </div>
            <div class="sc-event-edit-grid">
              <label class="sc-event-edit-field">
                <span>Type</span>
                <select v-model="editForm.optionType" class="form-select" :disabled="Boolean(editingTarget)" required>
                  <option value="kind">Kind</option>
                  <option value="fg">FG</option>
                  <option value="vessel">Vessel</option>
                </select>
              </label>
              <label class="sc-event-edit-field">
                <span>Status</span>
                <select v-model.number="editForm.isEnabled" class="form-select" required>
                  <option :value="1">启用</option>
                  <option :value="0">停用</option>
                </select>
              </label>
              <label class="sc-event-edit-field">
                <span>SortOrder</span>
                <input v-model.number="editForm.sortOrder" class="form-control" type="number" step="1" required />
              </label>
              <label class="sc-event-edit-field">
                <span>Name</span>
                <input v-model="editForm.optionName" class="form-control" type="text" required />
              </label>
            </div>
          </div>
          <div class="modal-footer-v2">
            <button type="button" class="button-secondary" :disabled="saving" @click="closeEdit">取消</button>
            <button type="submit" class="button-primary" :disabled="saving">{{ saving ? "保存中..." : "保存" }}</button>
          </div>
        </form>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="deleteModalOpen"
        class="modal-backdrop-v2 sc-event-delete-overlay"
        role="presentation"
        @click.self="closeDelete"
      >
        <section
          class="modal-panel-v2 sc-event-delete-panel"
          role="dialog"
          aria-modal="true"
          aria-labelledby="scOptionDeleteModalTitle"
          aria-describedby="scOptionDeleteDescription"
        >
          <div class="modal-header-v2">
            <h2 id="scOptionDeleteModalTitle">停用 SC Option</h2>
            <button type="button" class="modal-close-v2" aria-label="关闭" :disabled="deleting" @click="closeDelete">
              ×
            </button>
          </div>
          <div class="modal-body-v2">
            <p id="scOptionDeleteDescription" class="sc-event-delete-name">
              确认停用“<strong>{{ deletingItem?.optionName }}</strong
              >”吗？
            </p>
            <p class="sc-event-delete-warning">已有 SC Event 继续保留该选项，新建或修改其他事件时无法再选择它。</p>
            <div v-if="deleteMessage" class="modal-message-v2" role="alert">{{ deleteMessage }}</div>
          </div>
          <div class="modal-footer-v2">
            <button type="button" class="button-secondary" :disabled="deleting" @click="closeDelete">取消</button>
            <button
              type="button"
              class="button-secondary danger sc-event-delete-confirm"
              :disabled="deleting"
              @click="disableOption"
            >
              {{ deleting ? "停用中..." : "确认停用" }}
            </button>
          </div>
        </section>
      </div>
    </Teleport>
  </PageShell>
</template>

<script setup lang="ts">
import "@/features/crawler/styles/index.css";
import "@/features/sc/styles/index.css";
import { computed, reactive, ref } from "vue";
import { useRoute } from "vue-router";
import {
  createScOption,
  deleteScOption,
  getScOptions,
  updateScOption,
  type ScOptionGroups,
  type ScOptionItem,
  type ScOptionType,
  type ScOptionWriteInput
} from "@/features/sc/api";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LinkStrip from "@/shared/components/LinkStrip.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";

type OptionTypeFilter = "all" | ScOptionType;
type OptionStatusFilter = "all" | "enabled" | "disabled";
type ScOptionForm = ScOptionWriteInput & { optionType: ScOptionType };
type ScOptionIdentity = Pick<ScOptionItem, "id" | "optionType">;

const emptyGroups = (): ScOptionGroups => ({ kind: [], fg: [], vessel: [] });
const emptyForm = (): ScOptionForm => ({
  optionType: "kind",
  optionName: "",
  sortOrder: 0,
  isEnabled: 1
});

const route = useRoute();
const groups = ref<ScOptionGroups>(emptyGroups());
const loading = ref(false);
const error = ref("");
const pageMessage = ref("");
const editModalOpen = ref(false);
const editingTarget = ref<ScOptionIdentity | null>(null);
const editForm = reactive<ScOptionForm>(emptyForm());
const editMessage = ref("");
const editMessageOK = ref(false);
const saving = ref(false);
const deleteModalOpen = ref(false);
const deletingItem = ref<ScOptionItem | null>(null);
const deleteMessage = ref("");
const deleting = ref(false);

const typeFilters: Array<{ value: OptionTypeFilter; label: string }> = [
  { value: "all", label: "全部" },
  { value: "kind", label: "Kind" },
  { value: "fg", label: "FG" },
  { value: "vessel", label: "Vessel" }
];
const statusFilters: Array<{ value: OptionStatusFilter; label: string }> = [
  { value: "all", label: "全部" },
  { value: "enabled", label: "启用" },
  { value: "disabled", label: "停用" }
];

const typeFilter = computed<OptionTypeFilter>(() => normalizeTypeFilter(route.query.type));
const statusFilter = computed<OptionStatusFilter>(() => normalizeStatusFilter(route.query.status));
const typeLinks = computed(() =>
  typeFilters.map((item) => ({
    label: item.label,
    href: optionFilterHref(item.value, statusFilter.value),
    active: typeFilter.value === item.value
  }))
);
const statusLinks = computed(() =>
  statusFilters.map((item) => ({
    label: item.label,
    href: optionFilterHref(typeFilter.value, item.value),
    active: statusFilter.value === item.value
  }))
);
const items = computed(() => [...groups.value.kind, ...groups.value.fg, ...groups.value.vessel]);
const enabledCount = computed(() => items.value.filter((item) => item.isEnabled === 1).length);
const visibleItems = computed(() =>
  items.value.filter((item) => {
    if (typeFilter.value !== "all" && item.optionType !== typeFilter.value) return false;
    if (statusFilter.value === "enabled" && item.isEnabled !== 1) return false;
    if (statusFilter.value === "disabled" && item.isEnabled === 1) return false;
    return true;
  })
);

void load();

async function load() {
  loading.value = true;
  error.value = "";
  try {
    groups.value = await getScOptions(true);
  } catch (err) {
    error.value = errorText(err);
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editingTarget.value = null;
  Object.assign(editForm, emptyForm());
  editMessage.value = "";
  editModalOpen.value = true;
}

function openEdit(item: ScOptionItem) {
  editingTarget.value = { optionType: item.optionType, id: item.id };
  Object.assign(editForm, {
    optionType: item.optionType,
    optionName: item.optionName,
    sortOrder: item.sortOrder,
    isEnabled: item.isEnabled
  });
  editMessage.value = "";
  editModalOpen.value = true;
}

function closeEdit() {
  if (saving.value) return;
  editModalOpen.value = false;
}

async function saveOption() {
  saving.value = true;
  editMessage.value = "";
  try {
    const input: ScOptionWriteInput = {
      optionName: editForm.optionName,
      sortOrder: editForm.sortOrder,
      isEnabled: editForm.isEnabled
    };
    const target = editingTarget.value;
    if (target) await updateScOption(target.optionType, target.id, input);
    else await createScOption(editForm.optionType, input);
    pageMessage.value = target ? "SC Option 更新成功" : "SC Option 新增成功";
    editMessageOK.value = true;
    editModalOpen.value = false;
    await load();
  } catch (err) {
    editMessage.value = errorText(err);
    editMessageOK.value = false;
  } finally {
    saving.value = false;
  }
}

function openDelete(item: ScOptionItem) {
  deletingItem.value = item;
  deleteMessage.value = "";
  deleteModalOpen.value = true;
}

function closeDelete() {
  if (deleting.value) return;
  deleteModalOpen.value = false;
  deletingItem.value = null;
  deleteMessage.value = "";
}

async function disableOption() {
  const item = deletingItem.value;
  if (!item) return;
  deleting.value = true;
  deleteMessage.value = "";
  try {
    await deleteScOption(item.optionType, item.id);
    pageMessage.value = `SC Option“${item.optionName}”已停用`;
    deleteModalOpen.value = false;
    deletingItem.value = null;
    await load();
  } catch (err) {
    deleteMessage.value = errorText(err);
  } finally {
    deleting.value = false;
  }
}

function optionTypeLabel(optionType: ScOptionType): string {
  if (optionType === "kind") return "Kind";
  if (optionType === "fg") return "FG";
  return "Vessel";
}

function optionEventsHref(item: ScOptionItem): string {
  const query = new URLSearchParams();
  query.set(`${item.optionType}_option_id`, String(item.id));
  return `/scevents?${query.toString()}`;
}

function normalizeTypeFilter(value: unknown): OptionTypeFilter {
  const normalized = Array.isArray(value) ? value[0] : value;
  if (normalized === "kind" || normalized === "fg" || normalized === "vessel") return normalized;
  return "all";
}

function normalizeStatusFilter(value: unknown): OptionStatusFilter {
  const normalized = Array.isArray(value) ? value[0] : value;
  if (normalized === "enabled" || normalized === "disabled") return normalized;
  return "all";
}

function optionFilterHref(type: OptionTypeFilter, status: OptionStatusFilter): string {
  const query = new URLSearchParams();
  if (type !== "all") query.set("type", type);
  if (status !== "all") query.set("status", status);
  const encoded = query.toString();
  return encoded ? `/scoptionlist?${encoded}` : "/scoptionlist";
}

function formatUnixMinute(value: number): string {
  if (!value) return "-";
  const date = new Date(value * 1000);
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")} ${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function errorText(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}
</script>
