import { afterEach, describe, expect, it, vi } from "vitest";
import {
  createScOption,
  deleteScEvent,
  deleteScOption,
  getScOptions,
  runScPickV2,
  startScPickV2Copy,
  stopScPickV2Copy,
  updateScEvent,
  updateScOption
} from "./api";
import type { ScPickV2Request } from "./types";

describe("SC Event API", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("按事件名发送 DELETE 请求", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response('{"ok":true,"deleted":true}', {
        status: 200,
        headers: { "Content-Type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(deleteScEvent("2026-07-28 12/00")).resolves.toEqual({ ok: true, deleted: true });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/sc-events/2026-07-28%2012%2F00",
      expect.objectContaining({
        method: "DELETE",
        credentials: "same-origin"
      })
    );
  });

  it("读取并按分组归一化 ID 可重复的 SC 单选项", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          ok: true,
          kind: [
            {
              id: 1,
              optionName: "类型 A",
              sortOrder: 2,
              isEnabled: 1,
              scEventTimes: 12,
              lastScEventTime: 345,
              createdOn: 10,
              updatedOn: 20
            }
          ],
          fg: [{ id: 1, optionName: "FG B", sortOrder: 1, isEnabled: 0 }],
          vessel: []
        }),
        { status: 200, headers: { "Content-Type": "application/json" } }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(getScOptions()).resolves.toEqual({
      kind: [
        {
          id: 1,
          optionType: "kind",
          optionName: "类型 A",
          sortOrder: 2,
          isEnabled: 1,
          scEventTimes: 12,
          lastScEventTime: 345,
          createdOn: 10,
          updatedOn: 20
        }
      ],
      fg: [
        {
          id: 1,
          optionType: "fg",
          optionName: "FG B",
          sortOrder: 1,
          isEnabled: 0,
          scEventTimes: 0,
          lastScEventTime: 0,
          createdOn: 0,
          updatedOn: 0
        }
      ],
      vessel: []
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/sc-options", expect.objectContaining({ credentials: "same-origin" }));
  });

  it("读取包含停用项的 SC 管理列表", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response('{"ok":true,"kind":[],"fg":[],"vessel":[]}', {
        status: 200,
        headers: { "Content-Type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await getScOptions(true);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/sc-options?includeDisabled=1",
      expect.objectContaining({ credentials: "same-origin" })
    );
  });

  it("发送 SC 选项增改停用请求", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementation(() =>
        Promise.resolve(new Response('{"ok":true}', { status: 200, headers: { "Content-Type": "application/json" } }))
      );
    vi.stubGlobal("fetch", fetchMock);
    const input = {
      optionName: "类型 A",
      sortOrder: 2,
      isEnabled: 1
    };

    await createScOption("vessel", input);
    await updateScOption("kind", 7, input);
    await deleteScOption("fg", 7);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/sc-options/vessel",
      expect.objectContaining({ method: "POST", body: JSON.stringify(input) })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/sc-options/kind/7",
      expect.objectContaining({ method: "PUT", body: JSON.stringify(input) })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/sc-options/fg/7", expect.objectContaining({ method: "DELETE" }));
  });
});

describe("SC Event 图片编辑 API", () => {
  afterEach(() => vi.unstubAllGlobals());
  const input = {
    comeGListId: 0,
    movieCastPersonId: 0,
    kindOptionId: 0,
    duration: 12,
    fgOptionId: 0,
    vesselOptionId: 0,
    remarks: "事件备注"
  };

  it("选择图片时以 multipart 同时发送事件信息和图片", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response('{"ok":true,"item":{"ImagePath":"/images/event.png"}}'));
    vi.stubGlobal("fetch", fetchMock);
    const image = new File(["image content"], "event.png", { type: "image/png" });
    const response = await updateScEvent("event A/1", input, image);
    expect(response).toEqual({ ok: true, item: { ImagePath: "/images/event.png" } });
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/sc-events/event%20A%2F1");
    expect(init.method).toBe("POST");
    expect(init.credentials).toBe("same-origin");
    expect(new Headers(init.headers).has("Content-Type")).toBe(false);
    const data = init.body as FormData;
    expect(JSON.parse(String(data.get("data")))).toEqual(input);
    expect((data.get("image") as File).name).toBe("event.png");
    expect(await (data.get("image") as File).text()).toBe("image content");
  });

  it("普通信息编辑继续发送原 JSON 参数", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response('{"ok":true}'));
    vi.stubGlobal("fetch", fetchMock);
    await updateScEvent("event-a", input);
    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(init.body).toBe(JSON.stringify(input));
    expect(new Headers(init.headers).get("Content-Type")).toBe("application/json");
  });

  it("图片保存失败保留后端错误消息", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response('{"error":"保存图片失败"}', { status: 500 })));
    await expect(updateScEvent("event-a", input, new File(["image"], "event.png"))).rejects.toThrow("保存图片失败");
  });
});

describe("SC Pick V2 API", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("使用独立 V2 预览接口并保留标准卡片字段", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          preview_id: "preview-1",
          picked: 1,
          total_size_bytes: 1024,
          total_size_gb: "0.00",
          jav: {
            picked: 1,
            movies: [
              {
                name: "ABC-123",
                film_birth_date_w_media: "2026-01-02",
                video_url_w_media: "/media/abc.mp4",
                owned_w_media: 5,
                picked_order: 0
              }
            ],
            total_size_bytes: 1024,
            total_size_gb: "0.00",
            pick_info: { raw_candidate_movie_count: 5 }
          },
          fc2: { picked: 0, movies: [], total_size_bytes: 0, total_size_gb: "0.00", pick_info: {} }
        }),
        { status: 200, headers: { "Content-Type": "application/json" } }
      )
    );
    vi.stubGlobal("fetch", fetchMock);
    const request = {
      source: "wmedia",
      jav: { pickN: 1, options: {}, reqs: [{ weight: 1, req: {} }] },
      fc2: { pickN: 0, options: {}, reqs: [] }
    } as unknown as ScPickV2Request;

    const response = await runScPickV2(request);
    expect(response.previewID).toBe("preview-1");
    expect(response.jav.movies[0].video_url_w_media).toBe("/media/abc.mp4");
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/triggers/sc/v2/pick",
      expect.objectContaining({ method: "POST", body: JSON.stringify(request) })
    );
  });

  it("按任务 ID 启动和停止复制", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response('{"job_id":"job-1","state":"running","total":2,"done":0,"failures":[]}', {
          status: 200,
          headers: { "Content-Type": "application/json" }
        })
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    await startScPickV2Copy("preview-1");
    await stopScPickV2Copy("job-1");

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/triggers/sc/v2/copy",
      expect.objectContaining({ method: "POST", body: JSON.stringify({ preview_id: "preview-1" }) })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/triggers/sc/v2/copy/job-1/stop",
      expect.objectContaining({ method: "POST" })
    );
  });
});
