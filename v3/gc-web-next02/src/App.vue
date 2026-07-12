<template>
  <AppLayout>
    <RouterView />
  </AppLayout>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted } from "vue";
import { useRouter } from "vue-router";
import AppLayout from "@/layouts/AppLayout.vue";
import { compactSearchParams } from "@/utils/query";

const router = useRouter();

function handleGetFormSubmit(event: SubmitEvent) {
  if (event.defaultPrevented) {
    return;
  }
  const form = event.target instanceof HTMLFormElement ? event.target : null;
  if (!form || form.method.toLowerCase() !== "get" || form.target === "_blank") {
    return;
  }

  const action = new URL(form.action || window.location.href, window.location.origin);
  if (action.origin !== window.location.origin) {
    return;
  }

  const params = new URLSearchParams(action.search);
  const formData = new FormData(form);
  const formKeys = new Set<string>();
  formData.forEach((_value, key) => formKeys.add(key));
  formKeys.forEach((key) => params.delete(key));
  formData.forEach((value, key) => {
    if (typeof value === "string") {
      params.append(key, value);
    }
  });

  const compacted = compactSearchParams(action.pathname, params).toString();
  const href = `${action.pathname}${compacted ? `?${compacted}` : ""}${action.hash}`;
  event.preventDefault();
  void router.push(href);
}

onMounted(() => document.addEventListener("submit", handleGetFormSubmit));
onBeforeUnmount(() => document.removeEventListener("submit", handleGetFormSubmit));
</script>
