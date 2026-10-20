import { ref } from "vue";

export type MovieResourceTab = "javbus" | "sukebei" | "sehuatang";
export interface MovieResourceSort {
  key: string;
  order: "asc" | "desc";
}

export function useMovieResources() {
  const activeResourceTab = ref<MovieResourceTab>("javbus");
  const resourceSorts = ref<Record<string, MovieResourceSort>>({});
  const reset = (): void => {
    activeResourceTab.value = "javbus";
    resourceSorts.value = {};
  };
  return { activeResourceTab, resourceSorts, reset };
}
