import { ref } from "vue";
import { getSourceConfig, updateSourceConfig } from "@/features/dailybest/api";
import type { SourceConfig } from "@/features/dailybest/types";
import { errorMessage } from "@/shared/utils/format";

export function useSourceConfig() {
  const config = ref<SourceConfig | null>(null);
  const address = ref("");
  const loading = ref(false);
  const saving = ref(false);
  const loadError = ref("");
  const saveError = ref("");
  const saved = ref(false);

  async function loadConfig(): Promise<void> {
    loading.value = true;
    loadError.value = "";
    saveError.value = "";
    saved.value = false;
    try {
      config.value = await getSourceConfig();
      address.value = config.value.jav_address;
    } catch (reason) {
      loadError.value = errorMessage(reason);
    } finally {
      loading.value = false;
    }
  }

  async function saveConfig(): Promise<void> {
    if (saving.value) return;
    saving.value = true;
    saveError.value = "";
    saved.value = false;
    try {
      config.value = await updateSourceConfig(address.value.trim());
      address.value = config.value.jav_address;
      saved.value = true;
    } catch (reason) {
      saveError.value = errorMessage(reason);
    } finally {
      saving.value = false;
    }
  }

  return { config, address, loading, saving, loadError, saveError, saved, loadConfig, saveConfig };
}
