import assert from "node:assert/strict";
import { test } from "vitest";
import { ref } from "vue";
import { useAlbumSelection } from "./useAlbumSelection";

test("album selection handles row selection, select all and removed rows", () => {
  const rows = ref([{ id: 1 }, { id: 2 }]);
  const selection = useAlbumSelection(rows, (row) => row.id);
  selection.toggleRow(rows.value[0], true);
  assert.equal(selection.selectedCount.value, 1);
  assert.equal(selection.partialSelected.value, true);
  selection.toggleAll(true);
  assert.equal(selection.allSelected.value, true);
  selection.remove([1]);
  assert.deepEqual([...selection.selectedIDs.value], [2]);
  selection.clear();
  assert.equal(selection.selectedCount.value, 0);
});
