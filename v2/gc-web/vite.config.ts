import { fileURLToPath, URL } from "node:url";

import vue from "@vitejs/plugin-vue";
import { defineConfig } from "vite";

const apiTarget = "http://127.0.0.1:2041";

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    host: "0.0.0.0",
    port: 2040,
    proxy: {
      "/api/gc/v2": {
        target: apiTarget,
        changeOrigin: true,
      },
      "/Volumes/Expansion": {
        target: apiTarget,
        changeOrigin: true,
      },
      "/Volumes/Getea": {
        target: apiTarget,
        changeOrigin: true,
      },
      "/Volumes/movie-un": {
        target: apiTarget,
        changeOrigin: true,
      },
      "/Volumes/T7/data": {
        target: apiTarget,
        changeOrigin: true,
      },
    },
  },
});
