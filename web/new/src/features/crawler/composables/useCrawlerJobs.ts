import { onBeforeUnmount } from "vue";

export function useCrawlerJobs(refresh: () => void | Promise<void>) {
  let poller: number | undefined;

  function start(): void {
    if (poller !== undefined) return;
    poller = window.setInterval(() => void refresh(), 5000);
  }

  function stop(): void {
    if (poller === undefined) return;
    window.clearInterval(poller);
    poller = undefined;
  }

  onBeforeUnmount(stop);
  return { start, stop };
}
