import { readonly, ref } from "vue";

export type FeedbackToastTone = "success" | "error";

export interface FeedbackToastAction {
  label: string;
  href: string;
}

export interface FeedbackToast {
  id: number;
  message: string;
  tone: FeedbackToastTone;
  action?: FeedbackToastAction;
}

export interface ShowFeedbackToastOptions {
  tone?: FeedbackToastTone;
  action?: FeedbackToastAction;
}

const AUTO_DISMISS_MS = 5000;
const toasts = ref<FeedbackToast[]>([]);
const timers = new Map<number, ReturnType<typeof setTimeout>>();
let nextID = 1;

function dismissFeedbackToast(id: number): void {
  const timer = timers.get(id);
  if (timer !== undefined) {
    clearTimeout(timer);
    timers.delete(id);
  }
  toasts.value = toasts.value.filter((toast) => toast.id !== id);
}

function showFeedbackToast(message: string, options: ShowFeedbackToastOptions = {}): number | undefined {
  const normalizedMessage = message.trim();
  if (!normalizedMessage) return undefined;

  const id = nextID++;
  toasts.value = [
    ...toasts.value,
    {
      id,
      message: normalizedMessage,
      tone: options.tone ?? "success",
      action: options.action
    }
  ];
  timers.set(
    id,
    setTimeout(() => {
      dismissFeedbackToast(id);
    }, AUTO_DISMISS_MS)
  );
  return id;
}

function clearFeedbackToasts(): void {
  for (const timer of timers.values()) clearTimeout(timer);
  timers.clear();
  toasts.value = [];
}

export function useFeedbackToast() {
  return {
    toasts: readonly(toasts),
    show: showFeedbackToast,
    success: (message: string, action?: FeedbackToastAction) => showFeedbackToast(message, { tone: "success", action }),
    error: (message: string, action?: FeedbackToastAction) => showFeedbackToast(message, { tone: "error", action }),
    dismiss: dismissFeedbackToast,
    clear: clearFeedbackToasts
  };
}
