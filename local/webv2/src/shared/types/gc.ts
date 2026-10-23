export type AnyRecord = Record<string, unknown>;

export interface PageInfoLink {
  Label?: string;
  Href?: string;
  Active?: boolean;
  Disabled?: boolean;
  Ellipsis?: boolean;
}

export interface PageInfo {
  Total?: number;
  Page?: number;
  PageSize?: number;
  PageTotal?: number;
  StartHref?: string;
  PrevHref?: string;
  NextHref?: string;
  EndHref?: string;
  Links?: PageInfoLink[];
}

export interface NavLink {
  label: string;
  href: string;
}

export interface NavGroup {
  label: string;
  icon: string;
  links: NavLink[];
}
