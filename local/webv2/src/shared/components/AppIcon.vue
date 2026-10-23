<template>
  <svg
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.6"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <path :d="paths[name] || paths.photos" />
  </svg>
</template>
<script setup lang="ts">
defineProps<{ name: string }>();
const paths: Record<string, string> = {
  photos: "M4 4h6v6H4z M14 4h6v6h-6z M4 14h6v6H4z M14 14h6v6h-6z",
  time: "M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18 M12 7v5l3 2",
  albums: "M4 6h16v15H4z M7 3h10 M7 16l4-4 3 3 2-2 4 4 M8 10h.01",
  random: "M3 6h3l12 12h3 M18 15l3 3-3 3 M3 18h3l4-4 M14 10l4-4h3 M18 3l3 3-3 3",
  chart: "M4 3v18h17 M8 16v-5 M13 16V7 M18 16V4",
  capture: "M8 3H3v5 M16 3h5v5 M21 16v5h-5 M8 21H3v-5 M8 8h8v8H8z",
  tasks: "M9 5h12 M9 12h12 M9 19h12 M3 5l1 1 2-2 M3 12l1 1 2-2 M3 19l1 1 2-2",
  text: "M4 4h16v16H4z M7 8h10 M7 12h10 M7 16h6",
  mobile: "M7 2h10v20H7z M11 18h2",
  requests: "M4 5h16 M4 12h16 M4 19h16 M7 3v4 M17 10v4 M10 17v4",
  database: "M4 6c0-4 16-4 16 0s-16 4-16 0 M4 6v12c0 4 16 4 16 0V6 M4 12c0 4 16 4 16 0",
  lock: "M5 10h14v11H5z M8 10V6a4 4 0 0 1 8 0v4 M12 14v3",
  download: "M12 3v12 M7 10l5 5 5-5 M4 16v5h16v-5",
  archive: "M4 9h16v12H4z M3 3h18v5H3z M12 11v6 M9 14l3 3 3-3",
  unarchive: "M4 9h16v12H4z M3 3h18v5H3z M12 18v-6 M9 15l3-3 3 3",
  trash: "M4 7h16 M9 7V4h6v3 M7 7l1 14h8l1-14 M10 11v6 M14 11v6",
  palette:
    "M12 3a9 9 0 0 0 0 18h1.5a2 2 0 0 0 0-4H12a2 2 0 0 1 0-4h4a5 5 0 0 0 5-5c0-5-4-9-9-9 M7.5 9h.01 M10 6h.01 M15 6.5h.01 M17.5 10h.01"
};
</script>
