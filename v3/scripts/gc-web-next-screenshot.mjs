#!/usr/bin/env node
import {spawn} from "node:child_process";
import {mkdir, mkdtemp, readFile, rm, writeFile} from "node:fs/promises";
import {existsSync} from "node:fs";
import os from "node:os";
import path from "node:path";
import zlib from "node:zlib";

const chromePath = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const v3Root = path.resolve(path.dirname(new URL(import.meta.url).pathname), "..");
const dateArg = valueOf("--date") || new Date().toISOString().slice(0, 10);
const roundArg = valueOf("--round") || "01";
const roundName = `round-${String(roundArg).padStart(2, "0")}`;
const outRoot = path.join(v3Root, "tmp", "gc-web-next-screenshots", dateArg, roundName);
const viewport = {width: Number(valueOf("--width") || 1440), height: Number(valueOf("--height") || 1200)};
const base = {id: "gc-web", port: 5012, label: "v3/gc-web"};
const targets = [
  {id: "next02", port: 5022, label: "gc-web-next02"},
  {id: "next03", port: 5023, label: "gc-web-next03"},
  {id: "next04", port: 5024, label: "gc-web-next04"},
  {id: "next05", port: 5025, label: "gc-web-next05"}
];

function valueOf(name) {
  const index = process.argv.indexOf(name);
  return index >= 0 ? process.argv[index + 1] : "";
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function slugFor(value) {
  const slug = String(value || "root")
    .replace(/^\/+/, "")
    .replace(/\/+$/, "")
    .replaceAll("/", "__")
    .replace(/[^a-zA-Z0-9_.-]+/g, "-");
  return slug || "root";
}

async function fetchJSON(url) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`${url} ${response.status}`);
  }
  return response.json();
}

async function resolveSamples() {
  const samples = {
    movieName: "MCSR-639",
    castId: "229",
    scEventName: "2026-07-02-01-15",
    shThreadTid: "3607083"
  };
  try {
    const cards = await fetchJSON("http://127.0.0.1:5013/api/cards?ps=1");
    samples.movieName = cards.movies?.[0]?.name || samples.movieName;
  } catch (error) {
    console.warn(`sample movie fallback: ${error.message}`);
  }
  try {
    const casts = await fetchJSON("http://127.0.0.1:5013/api/casts?ps=1");
    samples.castId = String(casts.Items?.[0]?.Id || samples.castId);
  } catch (error) {
    console.warn(`sample cast fallback: ${error.message}`);
  }
  try {
    const sc = await fetchJSON("http://127.0.0.1:5013/api/sc-events?ps=1");
    samples.scEventName = sc.Items?.[0]?.Event?.Name || samples.scEventName;
  } catch (error) {
    console.warn(`sample sc fallback: ${error.message}`);
  }
  try {
    const sh = await fetchJSON("http://127.0.0.1:5013/api/fetch-site-sehuatang-sh-list?ps=1");
    samples.shThreadTid = String(sh.Rows?.[0]?.ThreadTID || samples.shThreadTid);
  } catch (error) {
    console.warn(`sample sh fallback: ${error.message}`);
  }
  return samples;
}

async function buildPages() {
  const samples = await resolveSamples();
  const moviePath = `/movie/${encodeURIComponent(samples.movieName)}`;
  const castPath = `/cast/${encodeURIComponent(samples.castId)}`;
  const scPath = `/sc-events/${encodeURIComponent(samples.scEventName)}`;
  const shDetailPath = `/fetch-site-sehuatang-sh-detail/${encodeURIComponent(samples.shThreadTid)}`;
  return [
    ["/", "Redirect", "root"],
    ["/records", "Redirect", "records"],
    ["/cards", "Cards"],
    ["/cardstoday", "Cards"],
    ["/cardshasrank", "Cards"],
    ["/cardsmediamowned", "Cards"],
    ["/cardsneeddownload", "Cards"],
    ["/cardsrandom", "Cards"],
    ["/moviecarddayrank", "Rank"],
    ["/moviecardperiodrank", "Rank"],
    [moviePath, "MovieDetail", `movie-${samples.movieName}`],
    ["/cast", "Cast"],
    [castPath, "Cast", `cast-${samples.castId}`],
    ["/medias", "List"],
    ["/crawl-records", "List"],
    ["/w-agg-events", "List"],
    ["/torrent-albums", "List"],
    ["/movie-albums", "List"],
    ["/d-seeds", "List"],
    ["/casts", "List"],
    ["/e-items", "List"],
    ["/cardsagg", "MovieAgg"],
    ["/movie-agg-all/release", "MovieAgg"],
    ["/movie-agg-all/release-bucket-list", "MovieAgg"],
    ["/w-media-agg/bucket-list", "MovieAgg"],
    ["/w-media-agg/birth", "MovieAgg"],
    ["/triggers/agg", "MovieAgg"],
    ["/wdir", "MovieAgg"],
    ["/sc-events", "SC"],
    ["/sc-events-cards", "SC"],
    [scPath, "SC", `sc-events-${samples.scEventName}`],
    ["/sc-pick-smart-media", "SC"],
    ["/triggers/sc", "SC"],
    ["/triggers/media", "SC"],
    ["/triggers/media-rescan", "SC"],
    ["/triggers/media-rollback", "SC"],
    ["/triggers/sc-media", "SC"],
    ["/triggers/dailybest", "Trigger"],
    ["/triggers/seeds", "Trigger"],
    ["/triggers/post-process", "Trigger"],
    ["/triggers/backfill", "Trigger"],
    ["/triggers/fetch-site", "Trigger"],
    ["/triggers/fetch-sehuatang", "Trigger"],
    ["/triggers/fetch-sehuatang-sh", "Trigger"],
    ["/triggers/fetch-sehuatang-sh-thread-list", "Trigger"],
    ["/triggers/fetch-sehuatang-sh-thread-detail", "Trigger"],
    ["/triggers/fetch-site-javbus-filtered", "Trigger"],
    ["/triggers/fetch-site-sukebei-filtered", "Trigger"],
    ["/fetch-sites", "FetchSite"],
    ["/fetch-site-javbus-list", "FetchSite"],
    ["/fetch-site-sukebei-list", "FetchSite"],
    ["/fetch-site-sehuatang-list", "FetchSite"],
    ["/fetch-site-sehuatang-sh-list", "FetchSite"],
    [shDetailPath, "FetchSite", `fetch-site-sehuatang-sh-detail-${samples.shThreadTid}`],
    ["/crawler/tasks", "Crawler"],
    ["/crawler/detail-loop", "Crawler"]
  ].map(([route, group, slug]) => ({route, group, slug: slug || slugFor(route)}));
}

async function freePort() {
  const net = await import("node:net");
  return new Promise((resolve, reject) => {
    const server = net.createServer();
    server.listen(0, "127.0.0.1", () => {
      const port = server.address().port;
      server.close(() => resolve(port));
    });
    server.on("error", reject);
  });
}

async function launchChrome() {
  if (!existsSync(chromePath)) {
    throw new Error(`Chrome not found: ${chromePath}`);
  }
  const port = await freePort();
  const userDataDir = await mkdtemp(path.join(os.tmpdir(), "gc-next-shot-"));
  const child = spawn(chromePath, [
    "--headless=new",
    "--disable-gpu",
    "--disable-dev-shm-usage",
    "--hide-scrollbars",
    "--no-first-run",
    "--no-default-browser-check",
    `--remote-debugging-port=${port}`,
    `--user-data-dir=${userDataDir}`,
    "about:blank"
  ], {stdio: ["ignore", "ignore", "pipe"]});
  child.stderr.on("data", (data) => {
    const text = String(data);
    if (!text.includes("DevTools listening")) {
      process.stderr.write(text);
    }
  });
  for (let i = 0; i < 80; i += 1) {
    try {
      const response = await fetch(`http://127.0.0.1:${port}/json/version`);
      if (response.ok) {
        return {child, port, userDataDir};
      }
    } catch {
      await sleep(100);
    }
  }
  child.kill("SIGTERM");
  throw new Error("Chrome DevTools did not start");
}

async function openTab(port) {
  const response = await fetch(`http://127.0.0.1:${port}/json/new`, {method: "PUT"});
  if (!response.ok) {
    throw new Error(`create tab failed: ${response.status}`);
  }
  const info = await response.json();
  return new CDP(info.webSocketDebuggerUrl);
}

class CDP {
  constructor(wsURL) {
    this.ws = new WebSocket(wsURL);
    this.nextId = 1;
    this.pending = new Map();
    this.waiters = new Map();
    this.opened = new Promise((resolve, reject) => {
      this.ws.addEventListener("open", resolve, {once: true});
      this.ws.addEventListener("error", reject, {once: true});
    });
    this.ws.addEventListener("message", (event) => this.onMessage(event));
  }

  onMessage(event) {
    const message = JSON.parse(String(event.data));
    if (message.id && this.pending.has(message.id)) {
      const {resolve, reject} = this.pending.get(message.id);
      this.pending.delete(message.id);
      if (message.error) {
        reject(new Error(message.error.message || JSON.stringify(message.error)));
      } else {
        resolve(message.result || {});
      }
      return;
    }
    if (message.method && this.waiters.has(message.method)) {
      const waiters = this.waiters.get(message.method);
      this.waiters.delete(message.method);
      for (const resolve of waiters) {
        resolve(message.params || {});
      }
    }
  }

  async send(method, params = {}) {
    await this.opened;
    const id = this.nextId;
    this.nextId += 1;
    this.ws.send(JSON.stringify({id, method, params}));
    return new Promise((resolve, reject) => {
      this.pending.set(id, {resolve, reject});
      setTimeout(() => {
        if (this.pending.has(id)) {
          this.pending.delete(id);
          reject(new Error(`${method} timeout`));
        }
      }, 30000);
    });
  }

  event(method, timeoutMs = 30000) {
    return new Promise((resolve, reject) => {
      const waiters = this.waiters.get(method) || [];
      waiters.push(resolve);
      this.waiters.set(method, waiters);
      setTimeout(() => reject(new Error(`${method} event timeout`)), timeoutMs);
    });
  }

  close() {
    this.ws.close();
  }
}

async function preparePage(session, url) {
  const load = session.event("Page.loadEventFired", 25000).catch(() => {});
  await session.send("Page.navigate", {url});
  await load;
  for (let i = 0; i < 80; i += 1) {
    const state = await session.send("Runtime.evaluate", {
      expression: `(() => ({ready: document.readyState, text: (document.body && document.body.innerText || "").length}))()`,
      returnByValue: true
    });
    const value = state.result?.value || {};
    if (value.ready === "complete" && value.text > 20) {
      break;
    }
    await sleep(250);
  }
  await session.send("Runtime.evaluate", {
    expression: `Promise.all(Array.from(document.images || []).filter((img) => !img.complete).map((img) => new Promise((resolve) => { const done = () => resolve(); img.onload = done; img.onerror = done; setTimeout(done, 1500); })))`,
    awaitPromise: true
  }).catch(() => {});
  await sleep(600);
}

async function capturePage(session, url, dir, slug) {
  await mkdir(dir, {recursive: true});
  await session.send("Emulation.setDeviceMetricsOverride", {
    width: viewport.width,
    height: viewport.height,
    deviceScaleFactor: 1,
    mobile: false
  });
  await preparePage(session, url);
  const metrics = await session.send("Page.getLayoutMetrics");
  const contentSize = metrics.contentSize || {};
  const contentHeight = Math.max(Math.ceil(contentSize.height || viewport.height), viewport.height);
  const contentWidth = Math.max(Math.ceil(contentSize.width || viewport.width), viewport.width);
  const meta = {url, contentWidth, contentHeight, viewport};
  meta.layoutBoxes = await collectLayoutBoxes(session);
  const fullPath = path.join(dir, `${slug}__full.png`);
  try {
    const full = await session.send("Page.captureScreenshot", {
      format: "png",
      captureBeyondViewport: true,
      fromSurface: true,
      clip: {x: 0, y: 0, width: viewport.width, height: contentHeight, scale: 1}
    });
    await writeFile(fullPath, Buffer.from(full.data, "base64"));
    meta.full = fullPath;
  } catch (error) {
    meta.fullError = error.message;
  }
  const partCount = Math.ceil(contentHeight / viewport.height);
  meta.parts = [];
  for (let index = 0; index < partCount; index += 1) {
    const y = index * viewport.height;
    const height = Math.min(viewport.height, contentHeight - y);
    const part = await session.send("Page.captureScreenshot", {
      format: "png",
      captureBeyondViewport: true,
      fromSurface: true,
      clip: {x: 0, y, width: viewport.width, height, scale: 1}
    });
    const partPath = path.join(dir, `${slug}__part-${String(index + 1).padStart(3, "0")}.png`);
    await writeFile(partPath, Buffer.from(part.data, "base64"));
    meta.parts.push(partPath);
  }
  return meta;
}

async function collectLayoutBoxes(session) {
  const expression = `(() => {
    const interestingTags = new Set(["A", "BUTTON", "IMG", "INPUT", "SELECT", "TEXTAREA", "TABLE", "THEAD", "TBODY", "TR", "TH", "TD", "ARTICLE", "SECTION", "NAV", "HEADER", "ASIDE", "MAIN", "FORM", "UL", "LI"]);
    const boxes = [];
    for (const el of Array.from(document.querySelectorAll("body *"))) {
      const style = getComputedStyle(el);
      if (style.display === "none" || style.visibility === "hidden" || Number(style.opacity || 1) === 0) continue;
      const rect = el.getBoundingClientRect();
      const width = Math.round(rect.width);
      const height = Math.round(rect.height);
      if (width < 8 || height < 8) continue;
      const area = width * height;
      if (area < 180) continue;
      const hasBoxStyle = style.backgroundColor !== "rgba(0, 0, 0, 0)" || style.borderTopWidth !== "0px" || style.boxShadow !== "none";
      const hasText = (el.innerText || "").trim().length > 0;
      const hasMedia = el.tagName === "IMG" || el.querySelector("img");
      if (!interestingTags.has(el.tagName) && !hasBoxStyle && !hasMedia && !hasText) continue;
      boxes.push({
        tag: el.tagName.toLowerCase(),
        x: Math.round(rect.left + window.scrollX),
        y: Math.round(rect.top + window.scrollY),
        w: width,
        h: height,
        area
      });
    }
    return boxes;
  })()`;
  const result = await session.send("Runtime.evaluate", {expression, returnByValue: true});
  return result.result?.value || [];
}

function parsePNG(buffer) {
  const signature = buffer.subarray(0, 8).toString("hex");
  if (signature !== "89504e470d0a1a0a") {
    throw new Error("not a png");
  }
  let offset = 8;
  let width = 0;
  let height = 0;
  let bitDepth = 0;
  let colorType = 0;
  const idat = [];
  while (offset < buffer.length) {
    const length = buffer.readUInt32BE(offset);
    const type = buffer.subarray(offset + 4, offset + 8).toString("ascii");
    const data = buffer.subarray(offset + 8, offset + 8 + length);
    offset += 12 + length;
    if (type === "IHDR") {
      width = data.readUInt32BE(0);
      height = data.readUInt32BE(4);
      bitDepth = data[8];
      colorType = data[9];
    } else if (type === "IDAT") {
      idat.push(data);
    } else if (type === "IEND") {
      break;
    }
  }
  if (bitDepth !== 8) {
    throw new Error(`unsupported png bit depth ${bitDepth}`);
  }
  const channels = colorType === 6 ? 4 : colorType === 2 ? 3 : colorType === 0 ? 1 : 0;
  if (!channels) {
    throw new Error(`unsupported png color type ${colorType}`);
  }
  const raw = zlib.inflateSync(Buffer.concat(idat));
  const stride = width * channels;
  const rgba = new Uint8ClampedArray(width * height * 4);
  let rawOffset = 0;
  let prev = Buffer.alloc(stride);
  for (let y = 0; y < height; y += 1) {
    const filter = raw[rawOffset];
    rawOffset += 1;
    const scan = Buffer.from(raw.subarray(rawOffset, rawOffset + stride));
    rawOffset += stride;
    for (let x = 0; x < stride; x += 1) {
      const left = x >= channels ? scan[x - channels] : 0;
      const up = prev[x] || 0;
      const upLeft = x >= channels ? prev[x - channels] || 0 : 0;
      let value = scan[x];
      if (filter === 1) value = (value + left) & 255;
      else if (filter === 2) value = (value + up) & 255;
      else if (filter === 3) value = (value + Math.floor((left + up) / 2)) & 255;
      else if (filter === 4) value = (value + paeth(left, up, upLeft)) & 255;
      scan[x] = value;
    }
    for (let x = 0; x < width; x += 1) {
      const source = x * channels;
      const target = (y * width + x) * 4;
      if (channels === 4) {
        rgba[target] = scan[source];
        rgba[target + 1] = scan[source + 1];
        rgba[target + 2] = scan[source + 2];
        rgba[target + 3] = scan[source + 3];
      } else if (channels === 3) {
        rgba[target] = scan[source];
        rgba[target + 1] = scan[source + 1];
        rgba[target + 2] = scan[source + 2];
        rgba[target + 3] = 255;
      } else {
        rgba[target] = scan[source];
        rgba[target + 1] = scan[source];
        rgba[target + 2] = scan[source];
        rgba[target + 3] = 255;
      }
    }
    prev = scan;
  }
  return {width, height, rgba};
}

function paeth(a, b, c) {
  const p = a + b - c;
  const pa = Math.abs(p - a);
  const pb = Math.abs(p - b);
  const pc = Math.abs(p - c);
  if (pa <= pb && pa <= pc) return a;
  if (pb <= pc) return b;
  return c;
}

function edgeMap(image, outW = 160, outH = 120) {
  const gray = new Uint8Array(outW * outH);
  for (let y = 0; y < outH; y += 1) {
    const sy = Math.min(image.height - 1, Math.floor((y + 0.5) * image.height / outH));
    for (let x = 0; x < outW; x += 1) {
      const sx = Math.min(image.width - 1, Math.floor((x + 0.5) * image.width / outW));
      const index = (sy * image.width + sx) * 4;
      gray[y * outW + x] = Math.round(image.rgba[index] * 0.299 + image.rgba[index + 1] * 0.587 + image.rgba[index + 2] * 0.114);
    }
  }
  const edges = new Uint8Array(outW * outH);
  for (let y = 0; y < outH - 1; y += 1) {
    for (let x = 0; x < outW - 1; x += 1) {
      const here = gray[y * outW + x];
      const dx = Math.abs(here - gray[y * outW + x + 1]);
      const dy = Math.abs(here - gray[(y + 1) * outW + x]);
      edges[y * outW + x] = dx + dy > 32 ? 1 : 0;
    }
  }
  return edges;
}

async function comparePNGs(basePath, targetPath) {
  const [baseImage, targetImage] = await Promise.all([
    readFile(basePath).then(parsePNG),
    readFile(targetPath).then(parsePNG)
  ]);
  const baseEdges = edgeMap(baseImage);
  const targetEdges = edgeMap(targetImage);
  let intersection = 0;
  let union = 0;
  let changed = 0;
  for (let i = 0; i < baseEdges.length; i += 1) {
    const a = baseEdges[i];
    const b = targetEdges[i];
    if (a || b) union += 1;
    if (a && b) intersection += 1;
    if (a !== b) changed += 1;
  }
  const edgeSimilarity = union ? intersection / union : 1;
  const binarySimilarity = 1 - changed / baseEdges.length;
  return {
    edgeSimilarity: round(edgeSimilarity * 100),
    binarySimilarity: round(binarySimilarity * 100),
    baseSize: `${baseImage.width}x${baseImage.height}`,
    targetSize: `${targetImage.width}x${targetImage.height}`
  };
}

function round(value) {
  return Math.round(value * 100) / 100;
}

async function comparePage(baseMeta, targetMeta) {
  const comparisons = [];
  const fullAvailable = baseMeta.full && targetMeta.full;
  if (fullAvailable) {
    comparisons.push({area: "full", ...(await comparePNGs(baseMeta.full, targetMeta.full)), layoutSimilarity: compareLayoutBoxes(baseMeta.layoutBoxes, targetMeta.layoutBoxes, 0, Math.max(baseMeta.contentHeight, targetMeta.contentHeight))});
  }
  const maxParts = Math.max(baseMeta.parts.length, targetMeta.parts.length);
  for (let i = 0; i < maxParts; i += 1) {
    const yStart = i * viewport.height;
    const yEnd = yStart + viewport.height;
    if (!baseMeta.parts[i] || !targetMeta.parts[i]) {
      comparisons.push({area: `part-${String(i + 1).padStart(3, "0")}`, edgeSimilarity: 0, binarySimilarity: 0, layoutSimilarity: 0, baseSize: baseMeta.parts[i] ? "present" : "missing", targetSize: targetMeta.parts[i] ? "present" : "missing"});
      continue;
    }
    comparisons.push({area: `part-${String(i + 1).padStart(3, "0")}`, ...(await comparePNGs(baseMeta.parts[i], targetMeta.parts[i])), layoutSimilarity: compareLayoutBoxes(baseMeta.layoutBoxes, targetMeta.layoutBoxes, yStart, yEnd)});
  }
  const partScores = comparisons.filter((item) => item.area !== "full").map((item) => item.layoutSimilarity);
  const averagePartSimilarity = partScores.length ? round(partScores.reduce((sum, value) => sum + value, 0) / partScores.length) : 0;
  return {averagePartSimilarity, comparisons};
}

function compareLayoutBoxes(baseBoxes, targetBoxes, yStart, yEnd) {
  const base = boxesInSlice(baseBoxes, yStart, yEnd);
  const target = boxesInSlice(targetBoxes, yStart, yEnd);
  if (!base.length && !target.length) return 100;
  if (!base.length || !target.length) return 0;
  const forward = averageBestIoU(base, target);
  const backward = averageBestIoU(target, base);
  const countPenalty = Math.min(base.length, target.length) / Math.max(base.length, target.length);
  return round(((forward + backward) / 2) * countPenalty * 100);
}

function boxesInSlice(boxes, yStart, yEnd) {
  return boxes
    .filter((box) => box.y < yEnd && box.y + box.h > yStart)
    .map((box) => ({
      x: box.x / viewport.width,
      y: (Math.max(box.y, yStart) - yStart) / viewport.height,
      w: box.w / viewport.width,
      h: (Math.min(box.y + box.h, yEnd) - Math.max(box.y, yStart)) / viewport.height,
      area: box.area
    }))
    .filter((box) => box.w > 0 && box.h > 0);
}

function averageBestIoU(source, target) {
  let sum = 0;
  for (const box of source) {
    let best = 0;
    for (const candidate of target) {
      best = Math.max(best, iou(box, candidate));
    }
    sum += best;
  }
  return source.length ? sum / source.length : 0;
}

function iou(a, b) {
  const left = Math.max(a.x, b.x);
  const top = Math.max(a.y, b.y);
  const right = Math.min(a.x + a.w, b.x + b.w);
  const bottom = Math.min(a.y + a.h, b.y + b.h);
  const intersection = Math.max(0, right - left) * Math.max(0, bottom - top);
  if (!intersection) return 0;
  const union = a.w * a.h + b.w * b.h - intersection;
  return union ? intersection / union : 0;
}

function pageURL(port, route) {
  return `http://127.0.0.1:${port}${route}`;
}

function renderReport(pages, results) {
  const lines = [];
  lines.push(`# gc-web-next 截图对比 ${dateArg} ${roundName}`);
  lines.push("");
  lines.push(`- 输出目录：\`${outRoot}\``);
  lines.push(`- 页面数量：${pages.length}`);
  lines.push(`- 版本范围：${targets.map((item) => item.id).join(", ")}`);
  lines.push(`- 截图方式：Chrome headless 整页截图 + ${viewport.width}x${viewport.height} 垂直分段截图`);
  lines.push(`- 相似度算法：可见元素几何布局相似度；截图继续保留用于人工复核`);
  lines.push("");
  lines.push("| 页面 | 分组 | 版本 | 分段数 | 平均布局相似度 | 结论 |");
  lines.push("| --- | --- | --- | --- | --- | --- |");
  for (const page of pages) {
    for (const target of targets) {
      const item = results.find((result) => result.slug === page.slug && result.target === target.id);
      const status = item && item.averagePartSimilarity > 95 ? "达标" : "未达标";
      lines.push(`| \`${page.route}\` | ${page.group} | ${target.id} | ${item?.partCount ?? "-"} | ${item?.averagePartSimilarity ?? "-"}% | ${status} |`);
    }
  }
  lines.push("");
  lines.push("## 截图路径");
  lines.push("");
  lines.push("| 页面 | 基准截图目录 | next02 | next03 | next04 | next05 |");
  lines.push("| --- | --- | --- | --- | --- | --- |");
  for (const page of pages) {
    const row = [
      `\`${page.route}\``,
      `\`${path.join(outRoot, base.id, page.slug)}\``,
      ...targets.map((target) => `\`${path.join(outRoot, target.id, page.slug)}\``)
    ];
    lines.push(`| ${row.join(" | ")} |`);
  }
  return `${lines.join("\n")}\n`;
}

async function main() {
  await mkdir(outRoot, {recursive: true});
  const pages = await buildPages();
  await writeFile(path.join(outRoot, "manifest.json"), JSON.stringify({date: dateArg, round: roundName, viewport, base, targets, pages}, null, 2));
  const chrome = await launchChrome();
  const session = await openTab(chrome.port);
  await session.send("Page.enable");
  await session.send("Runtime.enable");
  const captures = new Map();
  const results = [];
  try {
    for (const page of pages) {
      const baseDir = path.join(outRoot, base.id, page.slug);
      console.log(`[base] ${page.route}`);
      const baseMeta = await capturePage(session, pageURL(base.port, page.route), baseDir, page.slug);
      captures.set(`${base.id}:${page.slug}`, baseMeta);
    }
    for (const target of targets) {
      console.log(`[target-start] ${target.id}`);
      for (const page of pages) {
        const baseDir = path.join(outRoot, base.id, page.slug);
        const baseMeta = captures.get(`${base.id}:${page.slug}`);
        if (!baseMeta) {
          throw new Error(`missing base capture: ${page.route}`);
        }
        const targetDir = path.join(outRoot, target.id, page.slug);
        console.log(`[${target.id}] ${page.route}`);
        const targetMeta = await capturePage(session, pageURL(target.port, page.route), targetDir, page.slug);
        captures.set(`${target.id}:${page.slug}`, targetMeta);
        const comparison = await comparePage(baseMeta, targetMeta);
        results.push({
          route: page.route,
          slug: page.slug,
          group: page.group,
          target: target.id,
          partCount: targetMeta.parts.length,
          averagePartSimilarity: comparison.averagePartSimilarity,
          comparisons: comparison.comparisons,
          baseDir,
          targetDir
        });
        console.log(`[compare] ${target.id} ${page.route} ${comparison.averagePartSimilarity}%`);
      }
    }
  } finally {
    session.close();
    chrome.child.kill("SIGTERM");
    await rm(chrome.userDataDir, {recursive: true, force: true}).catch(() => {});
  }
  await writeFile(path.join(outRoot, "captures.json"), JSON.stringify(Object.fromEntries(captures), null, 2));
  await writeFile(path.join(outRoot, "results.json"), JSON.stringify(results, null, 2));
  await writeFile(path.join(outRoot, "report.md"), renderReport(pages, results));
  const failed = results.filter((item) => item.averagePartSimilarity <= 95).length;
  console.log(`done pages=${pages.length} comparisons=${results.length} failed=${failed}`);
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
