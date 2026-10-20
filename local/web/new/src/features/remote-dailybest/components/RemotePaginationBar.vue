<template>
  <div v-if="localPageInfo.Links?.length" class="remote-list-pager">
    <PaginationBar :page-info="localPageInfo" />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { compactPageQuery } from "@/features/remote-dailybest/pagination";
import type { PageInfo as RemotePageInfo } from "@/features/remote-dailybest/types";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import type { PageInfo, PageInfoLink } from "@/shared/types/gc";

const props = defineProps<{ pageInfo: RemotePageInfo }>();
const route = useRoute();
const router = useRouter();

const localPageInfo = computed<PageInfo>(() => {
  if (!props.pageInfo.total) return {};
  const pageTotal = Math.max(props.pageInfo.page_total, 1);
  const current = Math.min(Math.max(props.pageInfo.page, 1), pageTotal);
  const links: PageInfoLink[] = [
    pageLink("Start", 1, false, current === 1),
    pageLink("Prev", Math.max(current - 1, 1), false, current === 1)
  ];
  for (let page = 1; page <= pageTotal; page += 1) {
    links.push(pageLink(String(page), page, page === current));
  }
  links.push(
    pageLink("Next", Math.min(current + 1, pageTotal), false, current === pageTotal),
    pageLink("End", pageTotal, false, current === pageTotal)
  );
  return {
    Total: props.pageInfo.total,
    Page: current,
    PageSize: props.pageInfo.page_size,
    PageTotal: pageTotal,
    Links: links
  };
});

function pageLink(label: string, page: number, active = false, disabled = false): PageInfoLink {
  return {
    Label: label,
    Href: disabled
      ? undefined
      : router.resolve({
          path: route.path,
          query: compactPageQuery(route.query, page, props.pageInfo.page_size)
        }).href,
    Active: active,
    Disabled: disabled
  };
}
</script>
