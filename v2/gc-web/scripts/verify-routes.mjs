import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "..");
const router = readFileSync(resolve(root, "src/router/index.ts"), "utf8");
const routes = readFileSync(resolve(root, "src/data/routes.ts"), "utf8");

const failed = [];

const cardRoutes = [
  { path: "/cards", key: "cards" },
  { path: "/cardstoday", key: "cardstoday" },
  { path: "/cardshasrank", key: "cardshasrank" },
  { path: "/cardsmediamowned", key: "cardsmediamowned" },
  { path: "/cardsneeddownload", key: "cardsneeddownload" },
  { path: "/cardsrandom", key: "cardsrandom" },
];

for (const item of cardRoutes) {
  if (!routes.includes(`path: "${item.path}"`) || !routes.includes(`key: "${item.key}"`)) {
    failed.push(`${item.path} route is missing from route data`);
  }
}
if (!router.includes('redirect: "/cards"')) {
  failed.push("root redirect to /cards is missing");
}
if (!router.includes('path: "movie/:movie"') || !router.includes('name: "movie-detail"') || !router.includes('@/pages/movie/index.vue')) {
  failed.push("movie detail route is missing");
}
for (const token of ["cardRoutes.map", "cardView: item.key", '@/pages/cards/index.vue']) {
  if (!router.includes(token)) {
    failed.push(`cards route mapping missing ${token}`);
  }
}

if (failed.length > 0) {
  console.error(`Route checks failed:\n${failed.join("\n")}`);
  process.exit(1);
}

console.log("Verified cards route scope.");
