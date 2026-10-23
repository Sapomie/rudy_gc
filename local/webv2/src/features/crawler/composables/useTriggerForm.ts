import { reactive } from "vue";

export interface TriggerFormState {
  name: string;
  number: number;
  actorName: string;
  startPage: number;
  endPage: number;
  keyword: string;
  fetchMode: string;
  existingPolicy: string;
  postTimeFrom: string;
  postTimeTo: string;
  detailLimit?: number;
}

export function useTriggerForm(): TriggerFormState {
  return reactive<TriggerFormState>({
    name: "",
    number: 1000000,
    actorName: "",
    startPage: 0,
    endPage: 0,
    keyword: "",
    fetchMode: "search",
    existingPolicy: "skip",
    postTimeFrom: "",
    postTimeTo: "",
    detailLimit: undefined
  });
}
