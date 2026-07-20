export type HealthState = "loading" | "ready" | "unavailable";

export async function loadHealthState(): Promise<HealthState> {
  const [health, ready] = await Promise.all([fetch("/healthz"), fetch("/readyz")]);
  return health.ok && ready.ok ? "ready" : "unavailable";
}
