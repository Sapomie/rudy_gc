import { afterEach, describe, expect, it, vi } from "vitest";
import { createSSRApp } from "vue";
import { renderToString } from "@vue/server-renderer";
import SourceConfigPage from "./SourceConfigPage.vue";
import { useSourceConfig } from "../useSourceConfig";

vi.mock("@/features/remote-dailybest/useSourceConfig", async () => {
  const { ref } = await import("vue");
  const form = {
    config: ref({
      jav_address: "example.com",
      target_probe_url: "https://example.com/cn/vl_bestrated.php?&mode=&page=1",
      created_on: 10,
      updated_on: 20
    }),
    address: ref("example.com"),
    loading: ref(false),
    saving: ref(false),
    loadError: ref(""),
    saveError: ref(""),
    saved: ref(false),
    loadConfig: vi.fn(),
    saveConfig: vi.fn()
  };
  return { useSourceConfig: () => form };
});

const form = useSourceConfig();
afterEach(() => {
  form.loading.value = false;
  form.saving.value = false;
  form.loadError.value = "";
  form.saveError.value = "";
  form.saved.value = false;
});
async function renderPage() {
  return renderToString(createSSRApp(SourceConfigPage));
}

describe("站点配置页面", () => {
  it("沿用页面容器和表单样式，展示可编辑域名与当前探测地址", async () => {
    const html = await renderPage();
    expect(html).toContain("站点配置");
    expect(html).toContain("remote-dailybest-page");
    expect(html).toContain('value="example.com"');
    expect(html).toContain('aria-describedby="jav-address-hint"');
    expect(html).toContain("保存配置");
    expect(html).toContain("https://example.com/cn/vl_bestrated.php");
    expect(form.loadConfig).toHaveBeenCalled();
  });

  it("展示保存、失败、加载状态并禁用保存中的控件", async () => {
    form.saving.value = true;
    let html = await renderPage();
    expect(html).toContain("正在保存…");
    expect(html.match(/disabled/g)?.length).toBe(3);
    form.saving.value = false;
    form.saved.value = true;
    html = await renderPage();
    expect(html).toContain("保存成功");
    expect(html).toContain('role="status"');
    form.saved.value = false;
    form.saveError.value = "域名校验失败";
    expect(await renderPage()).toContain("域名校验失败");
    form.loadError.value = "配置读取失败";
    html = await renderPage();
    expect(html).toContain("重新加载");
    expect(html).toContain("配置读取失败");
    expect(html).not.toContain("保存配置");
    form.loading.value = true;
    expect(await renderPage()).toContain("loading-state");
  });
});
