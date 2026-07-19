export interface TriggerMetric {
  label: string;
  value: string | number;
  className?: string;
}

export interface TriggerMetricGroup {
  label: string;
  items: TriggerMetric[];
}

export interface TriggerEventEntry {
  key: string;
  level: "info" | "warn" | "error";
  html: string;
}
