import type { PageInfo, PageInfoLink } from "@/shared/types/gc";
import { normalizeHref } from "@/shared/utils/format";

export type SortDirection = "asc" | "desc";

export interface SortLinkModel {
  href: string;
  label: string;
  active?: boolean;
  direction?: SortDirection;
  disabled?: boolean;
}

export interface PaginationItem {
  key: string;
  label: string;
  href?: string;
  active: boolean;
  disabled: boolean;
  ellipsis: boolean;
}

type RawPaginationLink = PageInfoLink & {
  href?: string;
  label?: string;
  active?: boolean;
  disabled?: boolean;
  ellipsis?: boolean;
};

export function paginationItemsFromLinks(links: RawPaginationLink[] = []): PaginationItem[] {
  return links.map((link, index) => {
    const ellipsis = Boolean(link.Ellipsis ?? link.ellipsis);
    const disabled = Boolean(link.Disabled ?? link.disabled) || ellipsis;
    const label = ellipsis ? "..." : String(link.Label ?? link.label ?? "");
    const rawHref = link.Href ?? link.href;
    return {
      key: `${label || "ellipsis"}-${index}`,
      label,
      href: disabled || !rawHref ? undefined : normalizeHref(rawHref),
      active: Boolean(link.Active ?? link.active),
      disabled,
      ellipsis
    };
  });
}

export function paginationItems(pageInfo?: PageInfo): PaginationItem[] {
  return paginationItemsFromLinks(pageInfo?.Links || []);
}
