import { onBeforeUnmount, readonly, ref } from "vue";

export interface JobEventStreamHandlers {
  onMessage: (data: string) => void;
  onError: () => void;
}

export function useJobEventStream() {
  const activeKey = ref("");
  let source: EventSource | undefined;

  function stop(): void {
    source?.close();
    source = undefined;
    activeKey.value = "";
  }

  function connect(key: string, url: string, handlers: JobEventStreamHandlers): void {
    if (source && activeKey.value === key) return;
    stop();
    const nextSource = new EventSource(url);
    source = nextSource;
    activeKey.value = key;
    nextSource.onmessage = (event) => {
      if (source === nextSource) handlers.onMessage(event.data);
    };
    nextSource.onerror = () => {
      if (source !== nextSource) return;
      stop();
      handlers.onError();
    };
  }

  onBeforeUnmount(stop);
  return { activeKey: readonly(activeKey), connect, stop };
}
