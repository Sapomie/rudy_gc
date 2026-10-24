<template>
  <div class="style-library-view style-button-candidates">
    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>按钮候选</h2>
        <p>相同动作 · 相同尺寸 · 三种外观</p>
      </header>
      <div class="candidate-comparison">
        <article v-for="candidate in candidates" :key="candidate.id" class="candidate-panel">
          <header class="candidate-panel-heading">
            <span class="candidate-letter">{{ candidate.id.toUpperCase() }}</span>
            <div>
              <h3>{{ candidate.name }}</h3>
              <p>{{ candidate.description }}</p>
            </div>
          </header>
          <div class="candidate-sample-group">
            <h4>主要、普通与危险操作</h4>
            <div class="style-library-actions">
              <CandidateButton
                :candidate="candidate.id"
                tone="primary"
                :label="saving[candidate.id] ? '保存中…' : '保存'"
                :loading="saving[candidate.id]"
                @click="save(candidate.id)"
              />
              <CandidateButton :candidate="candidate.id" label="取消" @click="notify('取消')" />
              <CandidateButton :candidate="candidate.id" label="刷新" @click="notify('刷新')" />
              <CandidateButton :candidate="candidate.id" tone="danger" label="删除" @click="notify('删除')" />
            </div>
          </div>
          <div class="candidate-sample-group">
            <h4>图标 + 文字</h4>
            <div class="style-library-actions">
              <CandidateButton :candidate="candidate.id" label="下载封面" icon="download" @click="notify('下载封面')" />
              <CandidateButton :candidate="candidate.id" label="查看详情" icon="text" @click="notify('查看详情')" />
            </div>
          </div>
          <div class="candidate-sample-group">
            <h4>交互外观预览</h4>
            <div class="candidate-state-previews">
              <div v-for="state in interactionStates" :key="state.value">
                <CandidateButton
                  :candidate="candidate.id"
                  tone="primary"
                  label="保存"
                  :preview="state.value"
                  @click="save(candidate.id)"
                />
                <small>{{ state.label }}</small>
              </div>
            </div>
          </div>
          <div class="candidate-sample-group">
            <h4>不可用与处理中</h4>
            <div class="style-library-actions">
              <CandidateButton :candidate="candidate.id" tone="primary" label="保存" disabled />
              <CandidateButton :candidate="candidate.id" label="保存中…" loading />
            </div>
          </div>
        </article>
      </div>
      <div class="candidate-feedback" role="status" aria-live="polite">
        <span class="candidate-feedback-dot" aria-hidden="true" />
        {{ message || "点击按钮体验反馈；使用 Tab 查看键盘焦点。所有操作仅在此页演示。" }}
      </div>
    </section>

    <div class="style-library-columns">
      <section class="style-library-section">
        <header class="style-library-section-head">
          <h2>开关</h2>
          <p>用于启用或关闭一项设置</p>
        </header>
        <div class="style-library-section-body candidate-switch-list">
          <label
            v-for="setting in settings"
            :key="setting.id"
            class="candidate-switch-setting"
            :class="{ 'is-disabled': setting.disabled }"
          >
            <span class="candidate-setting-copy"
              ><strong>{{ setting.label }}</strong
              ><small>{{ setting.description }}</small></span
            >
            <span class="candidate-switch-value">{{ setting.enabled ? "已开启" : "已关闭" }}</span>
            <span class="candidate-switch">
              <input
                v-model="setting.enabled"
                type="checkbox"
                role="switch"
                :aria-label="setting.label"
                :disabled="setting.disabled"
              />
              <span class="candidate-switch-track" aria-hidden="true"><span /></span>
            </span>
          </label>
        </div>
      </section>
      <section class="style-library-section">
        <header class="style-library-section-head">
          <h2>选择控件</h2>
          <p>用于选择条件、数量与顺序</p>
        </header>
        <div class="style-library-section-body candidate-choice-list">
          <div class="candidate-choice-row">
            <span>拥有状态</span>
            <div class="link-strip candidate-segments" role="group" aria-label="拥有状态">
              <button
                v-for="option in ownershipOptions"
                :key="option.value"
                type="button"
                class="strip-link"
                :class="{ active: ownership === option.value }"
                :aria-pressed="ownership === option.value"
                @click="ownership = option.value"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
          <div class="candidate-choice-row">
            <span>每行卡片</span>
            <div class="card-grid-size-control candidate-segments">
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
          </div>
          <div class="candidate-choice-row">
            <span>排序</span>
            <div class="link-strip candidate-segments" aria-label="排序方式">
              <SortLink
                v-for="option in sortOptions"
                :key="option.value"
                href="#"
                class="strip-link"
                :label="option.label"
                :active="sort === option.value"
                :direction="descending ? 'desc' : 'asc'"
                @click.prevent="setSort(option.value)"
              />
            </div>
          </div>
          <p class="style-library-note" role="status">
            {{ ownershipOptions.find((option) => option.value === ownership)?.label }} · 每行 {{ columns }} 张 ·
            {{ sortOptions.find((option) => option.value === sort)?.label }}{{ descending ? "降序" : "升序" }}
          </p>
        </div>
      </section>
    </div>

    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>选择按钮方案</h2>
        <p>相同业务选项 · 五套独立结构、状态与交互</p>
      </header>
      <div class="style-library-section-body selection-candidate-showcase">
        <NativeRadioCandidate />
        <SegmentedChoiceCandidate />
        <PillChoiceCandidate />
        <CommandListCandidate />
        <ChoiceCardCandidate />
      </div>
    </section>

    <div class="style-library-columns">
      <section class="style-library-section">
        <header class="style-library-section-head">
          <h2>图标展示</h2>
          <p>图形与名称</p>
        </header>
        <div class="style-library-section-body candidate-icon-display">
          <div v-for="icon in displayIcons" :key="icon.name" class="candidate-static-icon">
            <AppIcon :name="icon.name" /><span>{{ icon.label }}</span>
          </div>
        </div>
      </section>
      <section class="style-library-section">
        <header class="style-library-section-head">
          <h2>纯图标操作</h2>
          <p>独立点击区域 · 悬停名称 · 键盘焦点</p>
        </header>
        <div class="style-library-section-body candidate-icon-actions">
          <div
            v-for="candidate in candidates"
            :key="candidate.id"
            class="candidate-icon-action-group"
            :aria-label="`方案 ${candidate.id.toUpperCase()} 图标操作`"
          >
            <span>{{ candidate.id.toUpperCase() }}</span>
            <CandidateButton
              :candidate="candidate.id"
              label="下载封面"
              icon="download"
              icon-only
              @click="notify('下载封面')"
            />
            <CandidateButton
              :candidate="candidate.id"
              label="查看详情"
              icon="text"
              icon-only
              @click="notify('查看详情')"
            />
            <CandidateButton
              :candidate="candidate.id"
              label="删除"
              icon="trash"
              tone="danger"
              icon-only
              @click="notify('删除')"
            />
          </div>
        </div>
      </section>
    </div>

    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>业务操作与状态</h2>
        <p>统一语义色 · 点击切换加入状态</p>
      </header>
      <div class="style-library-section-body candidate-business-grid">
        <div class="candidate-sample-group">
          <h3>播放</h3>
          <div class="candidate-business-row movie-card-media-block">
            <span>有字幕</span
            ><button type="button" class="btn btn-success" @click="notify('播放有字幕影片')">Play</button>
          </div>
          <div class="candidate-business-row movie-card-media-block">
            <span>无字幕</span
            ><button type="button" class="btn btn-info" @click="notify('播放无字幕影片')">Play</button>
          </div>
        </div>
        <div class="candidate-sample-group route-movie-detail">
          <h3>稍后下载</h3>
          <div v-for="item in laterSamples" :key="item.id" class="candidate-business-row">
            <span>{{ item.added ? "已加入" : "未加入" }}</span>
            <button
              type="button"
              class="btn movie-download-later-button"
              :class="item.added ? 'btn-success' : 'btn-outline-warning'"
              :aria-pressed="item.added"
              @click="item.added = !item.added"
            >
              {{ item.added ? "已添加" : "稍后下载" }}
            </button>
          </div>
        </div>
        <div class="candidate-sample-group route-sh-magnets">
          <h3>下载相册</h3>
          <div v-for="item in albumSamples" :key="item.id" class="candidate-business-row fetch-site-album-inline">
            <span>{{ item.added ? "已加入" : "未加入" }}</span>
            <button
              type="button"
              class="btn"
              :class="
                item.kind === 'pending'
                  ? item.added
                    ? 'btn-info'
                    : 'btn-outline-info'
                  : item.added
                    ? 'btn-warning'
                    : 'btn-outline-warning'
              "
              :aria-pressed="item.added"
              @click="item.added = !item.added"
            >
              {{
                item.kind === "pending" ? (item.added ? "已在待下载" : "待下载") : item.added ? "已在下载中" : "下载中"
              }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>表格中的按钮</h2>
        <p>30px 行高 · 24px 行内按钮</p>
      </header>
      <div class="table-scroll style-library-table-wrap">
        <table class="data-table table style-library-data-table candidate-table">
          <thead>
            <tr>
              <th>候选</th>
              <th>示例记录</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="candidate in candidates" :key="candidate.id">
              <td>{{ candidate.id.toUpperCase() }} · {{ candidate.name }}</td>
              <td>示例影片 {{ candidate.id.toUpperCase() }}</td>
              <td><StatusBadge label="已完成" tone="good" /></td>
              <td>
                <div class="style-library-actions">
                  <CandidateButton :candidate="candidate.id" label="查看详情" @click="notify('查看详情')" />
                  <CandidateButton
                    :candidate="candidate.id"
                    label="下载封面"
                    icon="download"
                    @click="notify('下载封面')"
                  />
                  <CandidateButton :candidate="candidate.id" label="删除" tone="danger" @click="notify('删除')" />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
    <p class="style-library-note">
      三组候选沿用粉色主题、统一圆角与间距；普通按钮 28px，表格内按钮 24px。候选仅用于样式库比较。
    </p>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref } from "vue";
import AppIcon from "@/shared/components/AppIcon.vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import StatusBadge from "@/features/remote-dailybest/components/StatusBadge.vue";
import CandidateButton from "../components/CandidateButton.vue";
import ChoiceCardCandidate from "../components/selection-candidates/ChoiceCardCandidate.vue";
import CommandListCandidate from "../components/selection-candidates/CommandListCandidate.vue";
import NativeRadioCandidate from "../components/selection-candidates/NativeRadioCandidate.vue";
import PillChoiceCandidate from "../components/selection-candidates/PillChoiceCandidate.vue";
import SegmentedChoiceCandidate from "../components/selection-candidates/SegmentedChoiceCandidate.vue";
import "../styles/button-candidates.css";

const candidates = [
  { id: "a", name: "清晰紧凑", description: "实心主操作 · 白底普通操作 · 红色描边危险操作" },
  { id: "b", name: "柔和浅底", description: "实心主操作 · 浅底普通操作 · 浅红危险操作" },
  { id: "c", name: "简洁描边", description: "粉色描边主操作 · 中性边框 · 红色描边危险操作" }
] as const;
const interactionStates = [
  { value: "hover", label: "悬停" },
  { value: "pressed", label: "按下" },
  { value: "focus", label: "键盘焦点" }
] as const;
const message = ref("");
let actionCount = 0;
const saving = ref({ a: false, b: false, c: false });
const saveTimers = new Set<ReturnType<typeof setTimeout>>();
function notify(action: string) {
  actionCount += 1;
  message.value = `${action} · 第 ${actionCount} 次演示，反馈已更新。`;
}
function save(candidate: "a" | "b" | "c") {
  if (saving.value[candidate]) return;
  saving.value[candidate] = true;
  message.value = `方案 ${candidate.toUpperCase()} 正在演示保存…`;
  const timer = setTimeout(() => {
    saving.value[candidate] = false;
    notify("保存完成");
    saveTimers.delete(timer);
  }, 900);
  saveTimers.add(timer);
}
onBeforeUnmount(() => saveTimers.forEach(clearTimeout));
const settings = ref([
  { id: "refresh", label: "自动刷新", description: "开启后自动更新列表", enabled: true, disabled: false },
  { id: "details", label: "显示详细信息", description: "开启后显示补充字段", enabled: false, disabled: false },
  { id: "locked", label: "锁定的设置", description: "不可编辑状态示例", enabled: true, disabled: true }
]);
const ownership = ref("all");
const columns = ref(6);
const ownershipOptions = [
  { value: "all", label: "全部" },
  { value: "owned", label: "已拥有" },
  { value: "missing", label: "未拥有" }
];
const sortOptions = [
  { value: "release", label: "上映日期" },
  { value: "rating", label: "评分" },
  { value: "download", label: "下载时间" }
];
const sort = ref("release");
const descending = ref(true);
function setSort(value: string) {
  descending.value = sort.value === value ? !descending.value : true;
  sort.value = value;
}
const displayIcons = [
  { name: "download", label: "下载" },
  { name: "text", label: "详情" },
  { name: "trash", label: "删除" },
  { name: "albums", label: "相册" }
];
const laterSamples = ref([
  { id: "later-off", added: false },
  { id: "later-on", added: true }
]);
const albumSamples = ref([
  { id: "pending-off", kind: "pending", added: false },
  { id: "pending-on", kind: "pending", added: true },
  { id: "downloading-off", kind: "downloading", added: false },
  { id: "downloading-on", kind: "downloading", added: true }
]);
</script>
