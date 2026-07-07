import http from "node:http";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const webRoot = __dirname;
const staticRoot = path.join(webRoot, "legacy-ui", "static");
const textRoot = path.resolve(webRoot, "..", "z_text");
const apiOrigin = process.env.V3_API_ORIGIN || "http://127.0.0.1:5013";
const port = Number(process.env.PORT || 5012);
const volumeRoots = [
  "/Volumes/Expansion",
  "/Volumes/Getea",
  "/Volumes/movie-un",
  "/Volumes/T7/data"
];
const cardsPagePaths = new Set([
  "/cards",
  "/cardstoday",
  "/cardshasrank",
  "/cardsmediamowned",
  "/cardsneeddownload",
  "/cardsrandom",
  "/moviecarddayrank",
  "/moviecardperiodrank"
]);
const movieAggPagePaths = new Set([
  "/cardsagg",
  "/movie-agg-all/release",
  "/movie-agg-all/release-bucket-list",
  "/w-media-agg/bucket-list",
  "/w-media-agg/birth",
  "/triggers/agg",
  "/wdir"
]);
const listPagePaths = new Set([
  "/medias",
  "/crawl-records",
  "/w-agg-events",
  "/torrent-albums",
  "/movie-albums",
  "/d-seeds",
  "/casts",
  "/e-items"
]);
const scPagePaths = new Set([
  "/sc-events",
  "/sc-events-cards",
  "/sc-pick-smart-media",
  "/triggers/sc",
  "/triggers/media",
  "/triggers/media-rescan",
  "/triggers/media-rollback",
  "/triggers/sc-media"
]);
const triggerPagePaths = new Set([
  "/triggers/dailybest",
  "/triggers/seeds",
  "/triggers/post-process",
  "/triggers/backfill",
  "/triggers/fetch-site",
  "/triggers/fetch-sehuatang",
  "/triggers/fetch-sehuatang-sh",
  "/triggers/fetch-sehuatang-sh-thread-list",
  "/triggers/fetch-sehuatang-sh-thread-detail",
  "/triggers/fetch-site-javbus-filtered",
  "/triggers/fetch-site-sukebei-filtered"
]);
const fetchSiteListPagePaths = new Set([
  "/fetch-site-javbus-list",
  "/fetch-site-sukebei-list",
  "/fetch-site-sehuatang-list",
  "/fetch-site-sehuatang-sh-list"
]);
const fetchSiteConfigPagePaths = new Set([
  "/fetch-sites"
]);

const contentTypes = {
  ".css": "text/css; charset=utf-8",
  ".html": "text/html; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".png": "image/png",
  ".jpg": "image/jpeg",
  ".jpeg": "image/jpeg",
  ".gif": "image/gif",
  ".svg": "image/svg+xml",
  ".mp4": "video/mp4",
  ".m4v": "video/mp4",
  ".mov": "video/quicktime",
  ".webm": "video/webm",
  ".mkv": "video/x-matroska",
  ".avi": "video/x-msvideo",
  ".ts": "video/mp2t",
  ".m2ts": "video/mp2t",
  ".woff": "font/woff",
  ".woff2": "font/woff2"
};

function contentTypeFor(filePath) {
  const ext = path.extname(filePath).toLowerCase();
  return contentTypes[ext] || "application/octet-stream";
}

function sendFile(res, filePath) {
  fs.readFile(filePath, (err, data) => {
    if (err) {
      res.writeHead(err.code === "ENOENT" ? 404 : 500);
      res.end(err.code === "ENOENT" ? "not found" : "server error");
      return;
    }
    res.writeHead(200, {"Content-Type": contentTypeFor(filePath)});
    res.end(data);
  });
}

function parseRangeHeader(rangeHeader, size) {
  const match = /^bytes=(\d*)-(\d*)$/.exec(String(rangeHeader || "").trim());
  if (!match) {
    return null;
  }
  const rawStart = match[1];
  const rawEnd = match[2];
  if (rawStart === "" && rawEnd === "") {
    return null;
  }

  let start;
  let end;
  if (rawStart === "") {
    const suffixLength = Number.parseInt(rawEnd, 10);
    if (!Number.isFinite(suffixLength) || suffixLength <= 0) {
      return null;
    }
    start = Math.max(size - suffixLength, 0);
    end = size - 1;
  } else {
    start = Number.parseInt(rawStart, 10);
    end = rawEnd === "" ? size - 1 : Number.parseInt(rawEnd, 10);
  }

  if (!Number.isFinite(start) || !Number.isFinite(end) || start < 0 || end < start || start >= size) {
    return null;
  }
  return {start, end: Math.min(end, size - 1)};
}

function sendStreamFile(req, res, filePath) {
  fs.stat(filePath, (err, stat) => {
    if (err) {
      res.writeHead(err.code === "ENOENT" ? 404 : 500);
      res.end(err.code === "ENOENT" ? "not found" : "server error");
      return;
    }
    if (!stat.isFile()) {
      res.writeHead(404);
      res.end("not found");
      return;
    }

    const size = stat.size;
    const contentType = contentTypeFor(filePath);
    const baseHeaders = {
      "Accept-Ranges": "bytes",
      "Content-Disposition": "inline",
      "Content-Type": contentType
    };
    const rangeHeader = req.headers.range;

    if (rangeHeader) {
      const range = parseRangeHeader(rangeHeader, size);
      if (!range) {
        res.writeHead(416, {
          ...baseHeaders,
          "Content-Range": `bytes */${size}`
        });
        res.end();
        return;
      }

      res.writeHead(206, {
        ...baseHeaders,
        "Content-Length": String(range.end - range.start + 1),
        "Content-Range": `bytes ${range.start}-${range.end}/${size}`
      });
      if (req.method === "HEAD") {
        res.end();
        return;
      }
      fs.createReadStream(filePath, {start: range.start, end: range.end}).pipe(res);
      return;
    }

    res.writeHead(200, {
      ...baseHeaders,
      "Content-Length": String(size)
    });
    if (req.method === "HEAD") {
      res.end();
      return;
    }
    fs.createReadStream(filePath).pipe(res);
  });
}

function safeJoin(root, requestPath) {
  const decoded = decodeURIComponent(requestPath);
  const target = path.normalize(path.join(root, decoded));
  if (!target.startsWith(root + path.sep) && target !== root) {
    return "";
  }
  return target;
}

function proxy(req, res) {
  const upstream = new URL(req.url || "/", apiOrigin);
  const proxyReq = http.request(upstream, {
    method: req.method,
    headers: {...req.headers, host: upstream.host}
  }, (proxyRes) => {
    res.writeHead(proxyRes.statusCode || 502, proxyRes.headers);
    proxyRes.pipe(res);
  });
  proxyReq.on("error", (err) => {
    res.writeHead(502, {"Content-Type": "application/json; charset=utf-8"});
    res.end(JSON.stringify({error: err.message}));
  });
  req.pipe(proxyReq);
}

function resolveVolumePath(requestPath) {
  const decoded = decodeURIComponent(requestPath);
  const normalized = path.normalize(decoded);
  for (const root of volumeRoots) {
    if (normalized === root || normalized.startsWith(root + path.sep)) {
      return normalized;
    }
  }
  return "";
}

http.createServer((req, res) => {
  const url = new URL(req.url || "/", `http://${req.headers.host || "127.0.0.1"}`);
  if (url.pathname === "/") {
    res.writeHead(302, {location: "/cards"});
    res.end();
    return;
  }
  if (cardsPagePaths.has(url.pathname)) {
    sendFile(res, path.join(webRoot, "cards.html"));
    return;
  }
  if (movieAggPagePaths.has(url.pathname)) {
    sendFile(res, path.join(webRoot, "movie-agg.html"));
    return;
  }
  if (listPagePaths.has(url.pathname)) {
    sendFile(res, path.join(webRoot, "list.html"));
    return;
  }
  if (url.pathname === "/records") {
    res.writeHead(302, {"Location": "/crawl-records"});
    res.end();
    return;
  }
  if (url.pathname === "/cast" || url.pathname.startsWith("/cast/")) {
    sendFile(res, path.join(webRoot, "cast-detail.html"));
    return;
  }
  if (scPagePaths.has(url.pathname) || url.pathname.startsWith("/sc-events/")) {
    sendFile(res, path.join(webRoot, "sc.html"));
    return;
  }
  if (triggerPagePaths.has(url.pathname)) {
    sendFile(res, path.join(webRoot, "triggers.html"));
    return;
  }
  if (fetchSiteListPagePaths.has(url.pathname)) {
    sendFile(res, path.join(webRoot, "fetch-site-list.html"));
    return;
  }
  if (url.pathname.startsWith("/fetch-site-sehuatang-sh-detail/")) {
    sendFile(res, path.join(webRoot, "sh-sehuatang-detail.html"));
    return;
  }
  if (fetchSiteConfigPagePaths.has(url.pathname)) {
    sendFile(res, path.join(webRoot, "fetch-sites.html"));
    return;
  }
  if (url.pathname === "/crawler/tasks") {
    sendFile(res, path.join(webRoot, "crawler-tasks.html"));
    return;
  }
  if (url.pathname === "/crawler/detail-loop") {
    sendFile(res, path.join(webRoot, "crawler-detail-loop.html"));
    return;
  }
  if (url.pathname.startsWith("/movie/")) {
    sendFile(res, path.join(webRoot, "movie-detail.html"));
    return;
  }
  if (url.pathname === "/favicon.ico") {
    res.writeHead(204);
    res.end();
    return;
  }
  if (url.pathname.startsWith("/api/")) {
    proxy(req, res);
    return;
  }
  if (url.pathname.startsWith("/Volumes/")) {
    const filePath = resolveVolumePath(url.pathname);
    if (!filePath) {
      res.writeHead(403);
      res.end("forbidden");
      return;
    }
    sendStreamFile(req, res, filePath);
    return;
  }
  if (url.pathname.startsWith("/text/")) {
    const filePath = safeJoin(textRoot, url.pathname.slice("/text/".length));
    if (!filePath) {
      res.writeHead(403);
      res.end("forbidden");
      return;
    }
    sendFile(res, filePath);
    return;
  }
  if (url.pathname.startsWith("/static/")) {
    const filePath = safeJoin(staticRoot, url.pathname.slice("/static/".length));
    if (!filePath) {
      res.writeHead(403);
      res.end("forbidden");
      return;
    }
    sendFile(res, filePath);
    return;
  }
  if (url.pathname.startsWith("/src/")) {
    const filePath = safeJoin(webRoot, url.pathname.slice(1));
    if (!filePath) {
      res.writeHead(403);
      res.end("forbidden");
      return;
    }
    sendFile(res, filePath);
    return;
  }
  res.writeHead(404);
  res.end("not found");
}).listen(port, () => {
  console.log(`v3 gc-web listening on http://127.0.0.1:${port}`);
  console.log(`proxying API to ${apiOrigin}`);
});
