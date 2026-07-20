import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

const apiOrigin = process.env.API_ORIGIN || "http://127.0.0.1:8891";
const devPort = Number(process.env.PORT || 8892);

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) }
  },
  server: {
    host: "0.0.0.0",
    port: devPort,
    strictPort: true,
    proxy: {
      "/healthz": { target: apiOrigin, changeOrigin: true },
      "/readyz": { target: apiOrigin, changeOrigin: true },
      "/api": { target: apiOrigin, changeOrigin: true }
    }
  },
  preview: { host: "0.0.0.0" }
});
