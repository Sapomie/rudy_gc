import http from "node:http";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const webRoot = __dirname;
const staticRoot = path.join(webRoot, "legacy-ui", "static");
const textRoot = path.resolve(webRoot, "..", "z_text");
const apiOrigin = process.env.V3_API_ORIGIN || "http://127.0.0.1:5013";
const port = Number(process.env.PORT || 3053);
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
  ".woff": "font/woff",
  ".woff2": "font/woff2"
};

function sendFile(res, filePath) {
  const ext = path.extname(filePath).toLowerCase();
  fs.readFile(filePath, (err, data) => {
    if (err) {
      res.writeHead(err.code === "ENOENT" ? 404 : 500);
      res.end(err.code === "ENOENT" ? "not found" : "server error");
      return;
    }
    res.writeHead(200, {"Content-Type": contentTypes[ext] || "application/octet-stream"});
    res.end(data);
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
    sendFile(res, filePath);
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
