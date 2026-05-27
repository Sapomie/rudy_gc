import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "..");
const files = {
  api: readFileSync(resolve(root, "src/api/gc.ts"), "utf8"),
  cards: readFileSync(resolve(root, "src/pages/cards/index.vue"), "utf8"),
  routes: readFileSync(resolve(root, "src/data/routes.ts"), "utf8"),
  tile: readFileSync(resolve(root, "src/components/MovieTile.vue"), "utf8"),
  movie: readFileSync(resolve(root, "src/pages/movie/index.vue"), "utf8"),
  css: readFileSync(resolve(root, "src/styles/index.css"), "utf8"),
};

const requiredFilters = [
  "wd",
  "cn",
  "pid",
  "gn",
  "dn",
  "pn",
  "mn",
  "ln",
  "lj",
  "an",
  "rs",
  "re",
  "mbs",
  "mbe",
  "cay",
  "cao",
  "srds",
  "srde",
  "drkmin",
  "nd",
  "mowned",
  "vwmin",
  "vwmax",
  "smin",
  "smax",
  "lsctmin",
  "lsctmax",
  "scmin",
  "scmax",
  "comin",
  "comax",
  "md1",
  "md2",
  "md3",
  "md4",
  "n",
];

const cardViews = ["cards", "cardstoday", "cardshasrank", "cardsmediamowned", "cardsneeddownload", "cardsrandom"];

const failed = [];

if (!files.api.includes('apiGet<CardsListResponse>("/cards", params)')) {
  failed.push("cards API wrapper missing");
}
if (!files.api.includes("apiGet<MovieDetailResponse>")) {
  failed.push("movie detail API wrapper missing");
}
if (!files.cards.includes("currentView") || !files.cards.includes("view: currentView.value")) {
  failed.push("cards page must pass current route view to API");
}
for (const view of cardViews) {
  if (!files.routes.includes(`key: "${view}"`)) {
    failed.push(`missing cards view route ${view}`);
  }
}
for (const key of requiredFilters) {
  if (!files.cards.includes(`key: "${key}"`)) {
    failed.push(`missing /cards filter ${key}`);
  }
}
for (const token of ["MovieGrid", "PaginationControl", "sortOptions", "ownedLinks", "filterOpen", "applyFilters", "goPage"]) {
  if (!files.cards.includes(token)) {
    failed.push(`missing cards behavior token ${token}`);
  }
}
for (const token of ["viewLinks", "currentPath", "routeQueryObject({ keepPaging: false })"]) {
  if (!files.cards.includes(token)) {
    failed.push(`missing cards multi-page behavior ${token}`);
  }
}
for (const token of ["movie.movie_href", "Number.isFinite(movie.score)", "Number.isFinite(movie.viewers_number_watched)", "movie.jacket_img", "MPlay", "MRemoved"]) {
  if (!files.tile.includes(token)) {
    failed.push(`missing movie tile contract ${token}`);
  }
}
for (const token of [
  "rgc-detail-hero",
  "rgc-detail-poster",
  "infoTab",
  "resourceTabs",
  "sortedJavbus",
  "sortedSukebei",
  "sortedSehuatang",
  "selectedMovieAlbums",
  "marked_delete",
  "source_torrent_hash",
]) {
  if (!files.movie.includes(token)) {
    failed.push(`missing movie detail behavior token ${token}`);
  }
}
for (const token of [".rgc-cards-toolbar", ".rgc-view-tabs", ".rgc-sort-strip", ".rgc-owned-bar", ".rgc-filter-panel", ".rgc-movie-grid"]) {
  if (!files.css.includes(token)) {
    failed.push(`cards styling missing ${token}`);
  }
}
for (const token of [".rgc-detail-hero", ".rgc-detail-info-panel", ".rgc-detail-data-table", ".rgc-detail-fetch-grid"]) {
  if (!files.css.includes(token)) {
    failed.push(`movie detail styling missing ${token}`);
  }
}
if (!files.css.includes("grid-template-columns: repeat(6")) {
  failed.push("cards grid styling missing");
}
if (Object.values(files).some((source) => source.includes("page-actions") || source.includes("runPageAction"))) {
  failed.push("generic page action code must not appear in /cards implementation");
}

if (failed.length > 0) {
  console.error(`Contract checks failed:\n${failed.join("\n")}`);
  process.exit(1);
}

console.log("Verified /cards and /movie detail API, page, route, styling, and no-fake-action contracts.");
