import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "..");

const files = [
  "src/App.vue",
  "src/layouts/AppLayout.vue",
  "src/pages/cards/index.vue",
  "src/pages/generic/index.vue",
  "src/pages/movie/index.vue",
  "src/pages/rank/day.vue",
  "src/pages/rank/period.vue",
  "src/pages/notfound/index.vue",
  "src/components/PageFrame.vue",
  "src/components/StateBlock.vue",
  "src/components/PaginationControl.vue",
  "src/components/MovieTile.vue",
  "src/components/MovieGrid.vue",
  "src/components/DataTable.vue",
  "src/main.ts",
  "src/router/index.ts",
  "src/api/request.ts",
  "src/api/gc.ts",
  "src/types/api.ts",
  "src/data/routes.ts",
  "src/styles/index.css",
];

const failed = [];

for (const file of files) {
  const source = readFileSync(resolve(root, file), "utf8");
  if (source.trim().length === 0) {
    failed.push(`${file}: empty`);
  }
  if (file.endsWith(".vue") && !source.includes("<template") && !source.includes("<script")) {
    failed.push(`${file}: invalid SFC`);
  }
}

if (failed.length > 0) {
  console.error(`Source checks failed:\n${failed.join("\n")}`);
  process.exit(1);
}

console.log(`Verified ${files.length} v2 source files.`);
