import fs from "node:fs";
import path from "node:path";
import { fileURLToPath, URL } from "node:url";
import type { IncomingMessage, ServerResponse } from "node:http";
import { defineConfig, type Plugin } from "vite";
import vue from "@vitejs/plugin-vue";

const apiOrigin = process.env.V3_API_ORIGIN || "http://127.0.0.1:5010";
const devPort = Number(process.env.PORT || 5022);

function serveLocalAsset(req: IncomingMessage, res: ServerResponse): boolean {
  const raw = req.url || "";
  const clean = decodeURIComponent(raw.split("?")[0] || "");
  if (!clean.startsWith("/Volumes/") && !clean.startsWith("/text/")) {
    return false;
  }
  const filePath = clean.startsWith("/text/") ? path.join("/text", clean.slice("/text/".length)) : clean;
  if (!fs.existsSync(filePath) || !fs.statSync(filePath).isFile()) {
    res.statusCode = 404;
    res.end("not found");
    return true;
  }
  const ext = path.extname(filePath).toLowerCase();
  const contentType =
    ext === ".jpg" || ext === ".jpeg" ? "image/jpeg" :
    ext === ".png" ? "image/png" :
    ext === ".webp" ? "image/webp" :
    ext === ".gif" ? "image/gif" :
    "application/octet-stream";
  res.setHeader("Content-Type", contentType);
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
    port: devPort,
    strictPort: true,
    proxy: {
      "/api": {
        target: apiOrigin,
        changeOrigin: true
      }
    }
  }
});
