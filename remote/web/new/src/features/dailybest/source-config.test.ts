import { afterEach, describe, expect, it, vi } from "vitest";
import { useSourceConfig } from "./useSourceConfig";
import { getSourceConfig, updateSourceConfig } from "./api";
import type { SourceConfig } from "./types";

const endpoint = "/api/v1/dailybest/source-config";
const oldConfig: SourceConfig = {
  jav_address: "old.example.com",
  target_probe_url: "https://old.example.com/cn/vl_bestrated.php?&mode=&page=1",
  created_on: 10,
  updated_on: 10
};
const newConfig: SourceConfig = {
  ...oldConfig,
  jav_address: "new.example.com",
  target_probe_url: "https://new.example.com/cn/vl_bestrated.php?&mode=&page=1",
  updated_on: 20
};
function response(payload: unknown, status = 200): Response {
  return new Response(JSON.stringify(payload), { status });
}
afterEach(() => vi.unstubAllGlobals());

describe("站点配置", () => {
  it("查询和保存都调用同一远端配置接口", async () => {
    const fetchMock = vi.fn().mockResolvedValueOnce(response(oldConfig)).mockResolvedValueOnce(response(newConfig));
    vi.stubGlobal("fetch", fetchMock);
    await expect(getSourceConfig()).resolves.toEqual(oldConfig);
    await expect(updateSourceConfig("new.example.com")).resolves.toEqual(newConfig);
    expect(fetchMock).toHaveBeenNthCalledWith(1, endpoint, expect.any(Object));
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      endpoint,
      expect.objectContaining({ method: "PUT", body: JSON.stringify({ jav_address: "new.example.com" }) })
    );
  });

  it("加载当前域名，保存后更新展示并能刷新读取", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(response(oldConfig))
      .mockResolvedValueOnce(response(newConfig))
      .mockResolvedValueOnce(response(newConfig));
    vi.stubGlobal("fetch", fetchMock);
    const form = useSourceConfig();
    const loading = form.loadConfig();
    expect(form.loading.value).toBe(true);
    await loading;
    expect(form.address.value).toBe("old.example.com");
    form.address.value = " new.example.com ";
    await form.saveConfig();
    expect(form.config.value).toEqual(newConfig);
    expect(form.address.value).toBe("new.example.com");
    expect(form.saved.value).toBe(true);
    expect(form.saving.value).toBe(false);
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      endpoint,
      expect.objectContaining({ body: JSON.stringify({ jav_address: "new.example.com" }) })
    );
    await form.loadConfig();
    expect(form.config.value).toEqual(newConfig);
    expect(form.saved.value).toBe(false);
  });

  it("保存失败展示远端校验信息并保留编辑内容和原配置", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValueOnce(response(oldConfig))
        .mockResolvedValueOnce(response({ code: "invalid_source_config", message: "JAV 域名格式无效" }, 400))
    );
    const form = useSourceConfig();
    await form.loadConfig();
    form.address.value = "invalid/domain";
    await form.saveConfig();
    expect(form.saveError.value).toBe("JAV 域名格式无效");
    expect(form.address.value).toBe("invalid/domain");
    expect(form.config.value).toEqual(oldConfig);
    expect(form.saved.value).toBe(false);
    expect(form.saving.value).toBe(false);
  });

  it("加载失败可重试，空配置可首次保存", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValueOnce(response({ message: "数据库读取失败" }, 500))
        .mockResolvedValueOnce(response({ jav_address: "", target_probe_url: "", created_on: 0, updated_on: 0 }))
        .mockResolvedValueOnce(response(newConfig))
    );
    const form = useSourceConfig();
    await form.loadConfig();
    expect(form.loadError.value).toBe("数据库读取失败");
    expect(form.loading.value).toBe(false);
    await form.loadConfig();
    expect(form.loadError.value).toBe("");
    expect(form.address.value).toBe("");
    form.address.value = "new.example.com";
    await form.saveConfig();
    expect(form.config.value).toEqual(newConfig);
  });

  it("保存处理中阻止重复提交", async () => {
    let finish!: (value: Response) => void;
    const fetchMock = vi.fn().mockImplementation(
      () =>
        new Promise<Response>((resolve) => {
          finish = resolve;
        })
    );
    vi.stubGlobal("fetch", fetchMock);
    const form = useSourceConfig();
    form.address.value = "new.example.com";
    const saving = form.saveConfig();
    expect(form.saving.value).toBe(true);
    await form.saveConfig();
    expect(fetchMock).toHaveBeenCalledTimes(1);
    finish(response(newConfig));
    await saving;
    expect(form.saving.value).toBe(false);
    expect(form.saved.value).toBe(true);
  });
});
