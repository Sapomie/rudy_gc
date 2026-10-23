import { reactive, ref } from "vue";
import type {
  ScPickV2FC2Options,
  ScPickV2Filter,
  ScPickV2Group,
  ScPickV2JAVOptions,
  ScPickV2Request
} from "@/features/sc/types";

type EditableJAVOptions = ScPickV2JAVOptions & { randomSeed: number | null };
type EditableFC2Options = ScPickV2FC2Options & { randomSeed: number | null };

let groupSequence = 0;

export function useScPickV2Form() {
  const javPickN = ref(32);
  const fc2PickN = ref(0);
  const javOptions = reactive<EditableJAVOptions>({
    castParticipatedScBlockDays: 30,
    castSelectedScBlockDays: 60,
    rank20Min: 0,
    rank100Min: 0,
    rank500Min: 0,
    castParticipatedScPenaltyAlpha: 0.7,
    castSelectedScPenaltyAlpha: 0.7,
    castOwnedScRatioPenaltyAlpha: 0,
    movieHasSelectedScPenaltyAlpha: 0.9,
    randomSeed: null
  });
  const fc2Options = reactive<EditableFC2Options>({
    movieParticipatedScBlockDays: 30,
    movieSelectedScBlockDays: 30,
    movieSelectedScPenaltyAlpha: 0.7,
    movieHasSelectedScPenaltyAlpha: 0.7,
    randomSeed: null
  });
  const javGroups = ref<ScPickV2Group[]>([newGroup("jav")]);
  const fc2Groups = ref<ScPickV2Group[]>([newGroup("fc2")]);

  function addGroup(type: "jav" | "fc2") {
    const target = type === "jav" ? javGroups : fc2Groups;
    target.value = [...target.value, newGroup(type)];
  }

  function updateGroup(type: "jav" | "fc2", group: ScPickV2Group) {
    const target = type === "jav" ? javGroups : fc2Groups;
    target.value = target.value.map((item) => (item.id === group.id ? group : item));
  }

  function removeGroup(type: "jav" | "fc2", groupID: string) {
    const target = type === "jav" ? javGroups : fc2Groups;
    if (target.value.length === 1) return;
    target.value = target.value.filter((group) => group.id !== groupID);
  }

  function buildRequest(): ScPickV2Request {
    validateNonNegativeInteger(javPickN.value, "JAV 数量");
    validateNonNegativeInteger(fc2PickN.value, "FC2 数量");
    if (javPickN.value === 0 && fc2PickN.value === 0) {
      throw new Error("JAV 与 FC2 至少需要一项数量大于 0");
    }
    validateOptions();
    return {
      source: "wmedia",
      jav: {
        pickN: javPickN.value,
        options: cleanOptions(javOptions),
        reqs: javPickN.value > 0 ? javGroups.value.map((group) => buildGroupRequest(group, "jav")) : []
      },
      fc2: {
        pickN: fc2PickN.value,
        options: cleanOptions(fc2Options),
        reqs: fc2PickN.value > 0 ? fc2Groups.value.map((group) => buildGroupRequest(group, "fc2")) : []
      }
    };
  }

  function validateOptions() {
    const ranks = [javOptions.rank20Min, javOptions.rank100Min, javOptions.rank500Min];
    for (const value of ranks) validateNonNegativeInteger(value, "JAV Rank 保底");
    if (ranks[0] > ranks[1] || ranks[1] > ranks[2] || ranks[2] > javPickN.value) {
      throw new Error("需要满足 0 ≤ Rank20 ≤ Rank100 ≤ Rank500 ≤ JAV 数量");
    }
    for (const [label, value] of Object.entries({
      "JAV 主演参与阻断天数": javOptions.castParticipatedScBlockDays,
      "JAV 主演入选阻断天数": javOptions.castSelectedScBlockDays,
      "FC2 影片参与阻断天数": fc2Options.movieParticipatedScBlockDays,
      "FC2 影片入选阻断天数": fc2Options.movieSelectedScBlockDays
    })) {
      validateNonNegativeInteger(value, label);
    }
    for (const [label, value] of Object.entries({
      "JAV 主演参与惩罚": javOptions.castParticipatedScPenaltyAlpha,
      "JAV 主演入选惩罚": javOptions.castSelectedScPenaltyAlpha,
      "JAV 主演入选比例惩罚": javOptions.castOwnedScRatioPenaltyAlpha,
      "JAV 影片已入选惩罚": javOptions.movieHasSelectedScPenaltyAlpha,
      "FC2 影片入选惩罚": fc2Options.movieSelectedScPenaltyAlpha,
      "FC2 影片已入选惩罚": fc2Options.movieHasSelectedScPenaltyAlpha
    })) {
      if (!Number.isFinite(value) || value < 0) throw new Error(`${label}需要填写非负数`);
    }
  }

  return {
    javPickN,
    fc2PickN,
    javOptions,
    fc2Options,
    javGroups,
    fc2Groups,
    addGroup,
    updateGroup,
    removeGroup,
    buildRequest
  };
}

function newGroup(type: "jav" | "fc2"): ScPickV2Group {
  groupSequence += 1;
  return {
    id: `${type}-${groupSequence}`,
    weight: 12,
    expanded: false,
    filter: type === "jav" ? { ReleasingDateStart: "", ReleasingDateEnd: "" } : { PostTimeStart: "", PostTimeEnd: "" }
  };
}

function buildGroupRequest(group: ScPickV2Group, type: "jav" | "fc2") {
  if (!Number.isInteger(group.weight) || group.weight <= 0) {
    throw new Error(`${type === "jav" ? "JAV" : "FC2"} 条件组权重需要填写正整数`);
  }
  const filter = cleanFilter(group.filter);
  if (type === "jav") {
    filter.Page = 1;
    filter.MediaOwned = 3;
    return { weight: group.weight, req: filter };
  }
  return { weight: group.weight, fc2Req: filter };
}

function cleanFilter(filter: ScPickV2Filter): ScPickV2Filter {
  const cleaned: ScPickV2Filter = {};
  for (const [key, value] of Object.entries(filter)) {
    if (value === null || value === "") continue;
    cleaned[key] = typeof value === "string" ? value.trim() : value;
  }
  return cleaned;
}

function cleanOptions<T extends Record<string, number | null>>(
  options: T
): Omit<T, "randomSeed"> & { randomSeed?: number } {
  const cleaned = { ...options } as Omit<T, "randomSeed"> & { randomSeed?: number };
  if (options.randomSeed === null) delete cleaned.randomSeed;
  return cleaned;
}

function validateNonNegativeInteger(value: number, label: string) {
  if (!Number.isInteger(value) || value < 0) throw new Error(`${label}需要填写非负整数`);
}
