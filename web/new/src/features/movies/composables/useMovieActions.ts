import { ref } from "vue";

export function useMovieActions() {
  const busy = ref(false);
  const message = ref("");
  const error = ref("");

  function reset(): void {
    busy.value = false;
    message.value = "";
    error.value = "";
  }

  async function execute<T>(action: () => Promise<T>, successMessage: string): Promise<T | undefined> {
    if (busy.value) return undefined;
    busy.value = true;
    message.value = "";
    error.value = "";
    try {
      const result = await action();
      message.value = successMessage;
      return result;
    } catch (cause) {
      error.value = cause instanceof Error ? cause.message : String(cause);
      return undefined;
    } finally {
      busy.value = false;
    }
  }

  return { busy, message, error, reset, execute };
}
