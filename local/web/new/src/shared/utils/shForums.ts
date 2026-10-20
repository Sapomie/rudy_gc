import { getJSON } from "@/shared/api/request";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

export interface SHForumOption {
  forumID: string;
  forumName: string;
  sortOrder: number;
}

export async function loadSHForumOptions(): Promise<SHForumOption[]> {
  const payload = asRecord(await getJSON("/api/sh-forums"));
  return asArray<unknown>(pick(payload, "items", "Items"))
    .map((item) => {
      const row = asRecord(item);
      return {
        forumID: text(row, "forum_id", "ForumID"),
        forumName: text(row, "forum_name", "ForumName"),
        sortOrder: Number(pick(row, "sort_order", "SortOrder") || 0)
      };
    })
    .filter((item) => item.forumID && item.forumName)
    .sort((left, right) => left.sortOrder - right.sortOrder || left.forumID.localeCompare(right.forumID));
}

export function shForumName(options: SHForumOption[], forumID: unknown): string {
  const value = String(forumID ?? "").trim();
  return options.find((item) => item.forumID === value)?.forumName || "未知版块";
}

export function shForumNames(options: SHForumOption[], forumIDs: unknown): string {
  const values = Array.isArray(forumIDs) ? forumIDs : String(forumIDs ?? "").split(",");
  const names = Array.from(
    new Set(
      values
        .map((item) => String(item ?? "").trim())
        .filter(Boolean)
        .map((item) => shForumName(options, item))
    )
  );
  return names.length ? names.join("、") : "未知版块";
}

export function shForumFilterOptions(options: SHForumOption[]): [string, string][] {
  return options.map((item) => [item.forumID, item.forumName]);
}
