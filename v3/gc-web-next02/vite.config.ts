import fs from "node:fs";
import path from "node:path";
import { fileURLToPath, URL } from "node:url";
import type { IncomingMessage, ServerResponse } from "node:http";
import { defineConfig, type Plugin } from "vite";
import vue from "@vitejs/plugin-vue";

const apiOrigin = process.env.V3_API_ORIGIN || "http://127.0.0.1:5010";
const devPort = Number(process.env.PORT || 5022);

const contentTypes: Record<string, string> = {
  ".jpg": "image/jpeg",
  ".jpeg": "image/jpeg",
  ".png": "image/png",
  ".webp": "image/webp",
  ".gif": "image/gif",
  ".mp4": "video/mp4",
  ".m4v": "video/mp4",
  ".webm": "video/webm",
  ".ogv": "video/ogg",
  ".ogg": "video/ogg",
  ".mov": "video/quicktime",
  ".mkv": "video/x-matroska",
  ".avi": "video/x-msvideo"
};

function serveLocalAsset(req: IncomingMessage, res: ServerResponse): boolean {
  const raw = req.url || "";
  const clean = decodeURIComponent(raw.split("?")[0] || "");
  if (!clean.startsWith("/Volumes/") && !clean.startsWith("/text/")) {
    return false;
  }
  const filePath = clean.startsWith("/text/") ? path.join("/text", clean.slice("/text/".length)) : clean;
  if (!fs.existsSync(filePath)) {
    res.statusCode = 404;
    res.end("not found");
    return true;
  }

  const stat = fs.statSync(filePath);
  if (!stat.isFile()) {
    res.statusCode = 404;
    res.end("not found");
    return true;
  }

  const ext = path.extname(filePath).toLowerCase();
  res.setHeader("Content-Type", contentTypes[ext] || "application/octet-stream");
  res.setHeader("Accept-Ranges", "bytes");

  const range = req.headers.range;
  if (range) {
    const match = /^bytes=(\d*)-(\d*)$/.exec(range);
    if (!match || (!match[1] && !match[2])) {
      res.statusCode = 416;
      res.setHeader("Content-Range", `bytes */${stat.size}`);
      res.end();
      return true;
    }

    const suffixLength = match[1] ? undefined : Number(match[2]);
    const start = suffixLength === undefined
      ? Number(match[1])
      : Math.max(stat.size - suffixLength, 0);
    const requestedEnd = match[2] && suffixLength === undefined
      ? Number(match[2])
      : stat.size - 1;
    const end = Math.min(requestedEnd, stat.size - 1);

    if (!Number.isSafeInteger(start) || !Number.isSafeInteger(end) || start < 0 || start > end || start >= stat.size) {
      res.statusCode = 416;
      res.setHeader("Content-Range", `bytes */${stat.size}`);
      res.end();
      return true;
    }

    res.statusCode = 206;
    res.setHeader("Content-Range", `bytes ${start}-${end}/${stat.size}`);
    res.setHeader("Content-Length", end - start + 1);
    if (req.method === "HEAD") {
      res.end();
      return true;
    }
    fs.createReadStream(filePath, { start, end }).pipe(res);
    return true;
  }

  res.statusCode = 200;
  res.setHeader("Content-Length", stat.size);
  if (req.method === "HEAD") {
    res.end();
    return true;
  }
  fs.createReadStream(filePath).pipe(res);
  return true;
}

function localAssetPlugin(): Plugin {
  return {
    name: "gc-web-next02-local-assets",
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        if (serveLocalAsset(req, res)) {
          return;
        }
        next();
      });
    }
  };
}

export default defineConfig({
  plugins: [vue(), localAssetPlugin()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url))
    }
  },
  server: {
    host: "0.0.0.0",
    port: devPort,
    strictPort: true,
    proxy: {
      "/api": {
        target: apiOrigin,
        changeOrigin: true
      }
    }
  },
  preview: {
    host: "0.0.0.0"
  }
});
