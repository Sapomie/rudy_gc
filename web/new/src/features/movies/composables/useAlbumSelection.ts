import { computed, ref, type Ref } from "vue";

export function useAlbumSelection<T>(rows: Ref<T[]>, rowID: (row: T) => number) {
  const selectedIDs = ref<Set<number>>(new Set());
  const selectedCount = computed(() => selectedIDs.value.size);
  const allSelected = computed(() => rows.value.length > 0 && selectedIDs.value.size === rows.value.length);
  const partialSelected = computed(() => selectedIDs.value.size > 0 && selectedIDs.value.size < rows.value.length);

  const isSelected = (row: T): boolean => selectedIDs.value.has(rowID(row));
  function toggleRow(row: T, checked: boolean): void {
    const next = new Set(selectedIDs.value);
    const id = rowID(row);
    if (checked) next.add(id);
    else next.delete(id);
    selectedIDs.value = next;
  }
  function toggleAll(checked: boolean): void {
    selectedIDs.value = checked ? new Set(rows.value.map(rowID).filter((id) => id > 0)) : new Set();
  }
  function clear(): void {
    selectedIDs.value = new Set();
  }
  function remove(ids: Iterable<number>): void {
    const removed = new Set(ids);
    selectedIDs.value = new Set([...selectedIDs.value].filter((id) => !removed.has(id)));
  }

  return { selectedIDs, selectedCount, allSelected, partialSelected, isSelected, toggleRow, toggleAll, clear, remove };
}
