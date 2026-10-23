import { computed, onMounted, readonly, ref, type CSSProperties } from "vue";

const storageKey = "rudy-gc:card-grid-columns";
const columns = ref<number | null>(null);
let restored = false;

export function parseCardGridColumns(value: unknown): number | null {
  if (typeof value !== "number" && typeof value !== "string") return null;
  const count = Number(value);
  return Number.isSafeInteger(count) && count > 0 ? count : null;
}

function restoreColumns() {
  if (restored || typeof window === "undefined") return;
  restored = true;
  try {
    columns.value = parseCardGridColumns(window.localStorage.getItem(storageKey));
  } catch {
    // 浏览器禁用存储时，仍可在当前页面调整布局。
  }
}

function setColumns(value: number | string) {
  const count = parseCardGridColumns(value);
  if (count === null) return;
  columns.value = count;
  try {
    window.localStorage.setItem(storageKey, String(count));
  } catch {
    // 存储失败不影响当前布局和页面之间的设置共享。
  }
}

function resetColumns() {
  columns.value = null;
  try {
    window.localStorage.removeItem(storageKey);
  } catch {
    // 恢复默认布局无需依赖浏览器存储。
  }
}

const gridStyle = computed<CSSProperties | undefined>(() =>
  columns.value === null
    ? undefined
    : { display: "grid", gridTemplateColumns: `repeat(${columns.value}, minmax(0, 1fr))` }
);
const gridItemStyle = computed<CSSProperties | undefined>(() =>
  columns.value === null ? undefined : { flex: "none", maxWidth: "none" }
);

export function useCardGridSize() {
  onMounted(restoreColumns);
  return { columns: readonly(columns), gridStyle, gridItemStyle, setColumns, resetColumns };
}
