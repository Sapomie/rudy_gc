import { createReadStream, existsSync, statSync } from "node:fs";
import { createServer, request as httpRequest } from "node:http";
import { request as httpsRequest } from "node:https";
import { extname, join, normalize, resolve } from "node:path";
import { URL } from "node:url";

const port = Number(process.env.PORT || 2040);
const apiPrefix = process.env.API_PREFIX || "/api/gc/v2";
const apiOrigin = process.env.API_ORIGIN || "http://gc-api:2041";
const proxyPrefixes = (process.env.PROXY_PREFIXES || "/Volumes/Expansion,/Volumes/Getea,/Volumes/movie-un,/Volumes/T7/data")
  .split(",")
  .map((item) => item.trim())
  .filter(Boolean);
const distRoot = resolve("/app/dist");

const contentTypes = {
  ".css": "text/css; charset=utf-8",
  ".gif": "image/gif",
  ".html": "text/html; charset=utf-8",
  ".ico": "image/x-icon",
  ".jpg": "image/jpeg",
  ".jpeg": "image/jpeg",
  ".js": "text/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".png": "image/png",
  ".svg": "image/svg+xml",
  ".txt": "text/plain; charset=utf-8",
  ".webp": "image/webp",
  ".woff": "font/woff",
  ".woff2": "font/woff2",
};

function proxyApi(req, res) {
  const upstream = new URL(req.url || "/", apiOrigin);
  const requestImpl = upstream.protocol === "https:" ? httpsRequest : httpRequest;
  const headers = { ...req.headers, host: upstream.host };
  const proxyReq = requestImpl(
    upstream,
    {
      method: req.method,
      headers,
    },
    (proxyRes) => {
      res.writeHead(proxyRes.statusCode || 502, proxyRes.headers);
      proxyRes.pipe(res);
    },
  );

  proxyReq.on("error", () => {
    res.writeHead(502, { "content-type": "text/plain; charset=utf-8" });
    res.end("Bad Gateway");
  });

  req.pipe(proxyReq);
}

function resolveStaticPath(rawUrl) {
  let parsed;
  try {
    parsed = new URL(rawUrl || "/", "http://localhost");
  } catch {
    return "";
  }

  const normalized = normalize(decodeURIComponent(parsed.pathname)).replace(/^(\.\.[/\\])+/, "");
  const filePath = resolve(join(distRoot, normalized));
  if (!filePath.startsWith(distRoot)) {
    return "";
  }
  if (existsSync(filePath) && statSync(filePath).isFile()) {
    return filePath;
  }
  return join(distRoot, "index.html");
}

function sendFile(req, res, filePath) {
  const ext = extname(filePath).toLowerCase();
  res.writeHead(200, {
    "content-type": contentTypes[ext] || "application/octet-stream",
  });
  if (req.method === "HEAD") {
    res.end();
    return;
  }
  createReadStream(filePath).pipe(res);
}

const server = createServer((req, res) => {
  const rawUrl = req.url || "";
  if (rawUrl.startsWith(apiPrefix) || proxyPrefixes.some((prefix) => rawUrl.startsWith(prefix))) {
    proxyApi(req, res);
    return;
  }
  if (req.method !== "GET" && req.method !== "HEAD") {
    res.writeHead(405, { "content-type": "text/plain; charset=utf-8" });
    res.end("Method Not Allowed");
    return;
  }

  const filePath = resolveStaticPath(rawUrl);
  if (!filePath || !existsSync(filePath)) {
    res.writeHead(404, { "content-type": "text/plain; charset=utf-8" });
    res.end("Not Found");
    return;
  }
  sendFile(req, res, filePath);
});

server.listen(port, "0.0.0.0");
