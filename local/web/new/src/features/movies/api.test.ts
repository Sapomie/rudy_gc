import { afterEach, describe, expect, it, vi } from "vitest";
import { listPersonCafos, updatePersonCafo, type CafoWriteInput } from "./api";

describe("人物 c_cafo 接口", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("读取人物关联的全部 c_cafo", async () => {
    const response = { ok: true, cafos: [{ id: 7, personId: 11, name: "Actor" }] };
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(response), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(listPersonCafos(11)).resolves.toEqual(response.cafos);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/persons/11/cafos",
      expect.objectContaining({ credentials: "same-origin" })
    );
  });

  it("按人物与 cafo 双重标识提交编辑资料", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response('{"ok":true,"message":"演员资料更新成功"}', {
        status: 200,
        headers: { "Content-Type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);
    const input: CafoWriteInput = {
      name: "Actor",
      alias: "Alias",
      chinese: "中文名",
      birthDay: 946684800,
      height: 168,
      cup: "C",
      bwh: "90-60-88",
      avatar: "/images/actor.jpg",
      biography: "人物简介"
    };

    await updatePersonCafo(11, 7, input);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/persons/11/cafos/7",
      expect.objectContaining({ method: "PUT", body: JSON.stringify(input), credentials: "same-origin" })
    );
  });
});
