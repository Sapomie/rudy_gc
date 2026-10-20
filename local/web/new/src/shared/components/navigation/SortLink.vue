<template>
  <a
    :href="resolvedDisabled ? undefined : resolvedHref"
    class="sort-link"
    :class="[
      { active: resolvedActive, disabled: resolvedDisabled, desc: resolvedDirection === 'desc' },
      resolvedActive ? activeClass : ''
    ]"
    :aria-current="resolvedActive ? 'page' : undefined"
    :aria-disabled="resolvedDisabled || undefined"
  >
    <slot>{{ resolvedLabel }}</slot>
    <span v-if="resolvedActive && resolvedDirection" class="sort-direction" :class="indicatorClass" aria-hidden="true">
      {{ resolvedDirection === "desc" ? "↓" : "↑" }}
    </span>
  </a>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { SortDirection } from "./models";
import { normalizeHref } from "@/shared/utils/format";

type RawSortLink = Record<string, unknown>;

const props = withDefaults(
  defineProps<{
    href?: string;
    link?: RawSortLink;
    label?: string;
    active?: boolean;
    direction?: SortDirection;
    disabled?: boolean;
    indicatorClass?: string;
    activeClass?: string;
  }>(),
  {
    href: undefined,
    link: undefined,
    label: "",
    active: false,
    direction: undefined,
    disabled: false,
    indicatorClass: "",
    activeClass: ""
  }
);

const resolvedHref = computed(() => normalizeHref(String(props.href ?? props.link?.Href ?? props.link?.href ?? "#")));
const resolvedLabel = computed(() => props.label || String(props.link?.Label ?? props.link?.label ?? ""));
const resolvedActive = computed(() => props.active || Boolean(props.link?.Active ?? props.link?.active));
const resolvedDisabled = computed(() => props.disabled || Boolean(props.link?.Disabled ?? props.link?.disabled));
const resolvedDirection = computed<SortDirection | undefined>(() => {
  if (props.direction) return props.direction;
  if (!resolvedActive.value) return undefined;
  return (props.link?.Desc ?? props.link?.desc) ? "desc" : "asc";
});
</script>
