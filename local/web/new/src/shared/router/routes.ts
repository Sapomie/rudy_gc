import type { RouteRecordRaw } from "vue-router";

export interface AppRouteDefinition {
  path: string;
  name: string;
  title: string;
  apiPath?: string;
  component: RouteRecordRaw["component"];
}

export function defineFeatureRoutes(definitions: AppRouteDefinition[]): RouteRecordRaw[] {
  return definitions.map(({ path, name, title, apiPath, component }) => ({
    path,
    name,
    component,
    meta: { title, apiPath }
  }));
}
