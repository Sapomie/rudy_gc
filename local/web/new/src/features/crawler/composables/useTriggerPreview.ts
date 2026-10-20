import { ref } from "vue";

export function useTriggerPreview() {
  const open = ref(false);
  const loading = ref(false);
  const error = ref("");

  function begin(): void {
    open.value = true;
    loading.value = true;
    error.value = "";
  }

  function finish(message = ""): void {
    loading.value = false;
    error.value = message;
  }

  function reset(): void {
    open.value = false;
    loading.value = false;
    error.value = "";
  }

  return { open, loading, error, begin, finish, reset };
}
