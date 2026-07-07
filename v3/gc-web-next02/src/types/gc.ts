export type AnyRecord = Record<string, unknown>;

export interface PageInfoLink {
  Label?: string;
  Href?: string;
  Active?: boolean;
  Disabled?: boolean;
  Ellipsis?: boolean;
}

export interface PageInfo {
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
