import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import { useScPickV2Form } from "@/features/sc/composables/useScPickV2Form";

const resultsComponentSource = readFileSync(new URL("./components/pick-v2/ScPickResults.vue", import.meta.url), "utf8");

describe("SC Pick V2 form", () => {
  it("使用响应式状态构建 JAV 和 FC2 分组请求", () => {
    const form = useScPickV2Form();
    form.javPickN.value = 2;
    form.fc2PickN.value = 1;
    form.javOptions.rank100Min = 1;
    form.javOptions.rank500Min = 1;
    form.javGroups.value[0].filter.CastNames = "演员甲";
    form.fc2Groups.value[0].filter.HasSub = 2;

    const request = form.buildRequest();
    expect(request.jav.reqs[0]).toEqual({
      weight: 12,
      req: { Page: 1, MediaOwned: 3, CastNames: "演员甲" }
    });
    expect(request.fc2.reqs[0]).toEqual({ weight: 12, fc2Req: { HasSub: 2 } });
    expect(request.jav.options.movieHasSelectedScPenaltyAlpha).toBe(0.9);
    expect(request.fc2.options.movieHasSelectedScPenaltyAlpha).toBe(0.7);
  });

  it("拒绝超过 JAV 抽取数量的累计排名保底", () => {
    const form = useScPickV2Form();
    form.javPickN.value = 2;
    form.javOptions.rank500Min = 3;
    expect(() => form.buildRequest()).toThrow("Rank500");
  });

  it("JAV 结果默认按上映日期倒序展示", () => {
    expect(resultsComponentSource).toContain('const javSort = ref("release");');
    expect(resultsComponentSource).toContain("const javAscending = ref(false);");
  });
});
