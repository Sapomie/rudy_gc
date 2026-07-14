export type CompactQuery = Record<string, string | string[]>;

type QueryDefaults = Record<string, string>;
type QueryInput = Record<string, unknown>;

const listDefaults = (pageSize: string, sort: string, order = "desc"): QueryDefaults => ({
  ps: pageSize,
  sort,
  order
});

function movieCardDefaults(orderBy: string): QueryDefaults {
  return {
    ps: "18",
    od: orderBy,
    mowned: "1"
  };
}

function routeDefaults(path: string, query: QueryInput): QueryDefaults {
  switch (path) {
    case "/fclistcard":
    case "/fclistcardowned":
    case "/fclistcardrandom":
      return { ps: "18" };
    case "/listcard":
    case "/listcardtoday":
    case "/listcardneed":
    case "/listcardrandom":
    case "/listcardagg":
      return movieCardDefaults("releasing_date");
    case "/listcardrank":
      return movieCardDefaults("rank_date");
    case "/listcardowned":
      return movieCardDefaults("media_birth_time");
    case "/listcarddayrank":
      return { ps: "18" };
    case "/listcardperiodrank":
      return { ps: "18", type: "month", category: "1" };
    case "/aggrelease":
      return {
        ps: "100",
        od: "releasing_date",
        agg_mode: "all",
        year: "0",
        quarter: "0",
        month: "0",
        day: "0",
        tn: "30",
        top: "30",
        topn: "30",
        tc: "30"
      };
    case "/aggbirth":
      return {
        ps: "100",
        od: "media_birth_time",
        year: "0",
        quarter: "0",
        month: "0",
        day: "0",
        tn: "30",
        top: "30",
        topn: "30",
        tc: "30"
      };
    case "/bucketsrelease":
      return { ps: "100", sort: "scope", dir: "desc", level: "month", agg_mode: "all" };
    case "/bucketsbirth":
      return { ps: "100", sort: "scope", dir: "desc", level: "month" };
    case "/mediadir": {
      const rawID = firstQueryValue(query.id);
      if (!rawID || rawID === "root") {
        return { ps: "60", id: "root" };
      }
      return { ps: "18", od: "media_birth_time", recursive: "true", cp: "1", cs: "60" };
    }
    case "/scevents":
      return listDefaults("24", "sc_time");
    case "/sccards":
      return listDefaults("18", "sc_time");
    case "/crawlrecordlist":
      return { ...listDefaults("50", "start_time"), type: "SeedsActive" };
    case "/medialist":
      return listDefaults("24", "birth_time");
    case "/aggeventlist":
      return listDefaults("50", "started_time");
    case "/torrentalbumlist":
    case "/moviealbumlist":
      return listDefaults("50", "created_on");
    case "/seedlist":
      return listDefaults("50", "updated_on");
    case "/castlist":
      return listDefaults("24", "owned_w_media_number");
    case "/eitemlist":
      return listDefaults("50", "last_query_detail_time");
    case "/javbuslist":
    case "/sukebeilist":
      return listDefaults("50", "last_fetch_time");
    case "/sehuatanglist":
    case "/shpool":
      return listDefaults("50", "post_time");
    default:
      return {};
  }
}

function queryValues(value: unknown): string[] {
  const values = Array.isArray(value) ? value : [value];
  return values
    .filter((item) => item !== undefined && item !== null)
    .map((item) => String(item))
    .filter((item) => item.trim() !== "");
}

function firstQueryValue(value: unknown): string {
  return queryValues(value)[0] || "";
}

export function compactRouteQuery(path: string, query: QueryInput): CompactQuery {
  const defaults = routeDefaults(path, query);
  const compacted: CompactQuery = {};

  for (const [key, rawValue] of Object.entries(query)) {
    const values = queryValues(rawValue);
    if (!values.length) {
      continue;
    }
    if (key === "p" && values.length === 1 && values[0] === "1") {
      continue;
    }
    if (values.length === 1 && defaults[key] === values[0]) {
      continue;
    }
    compacted[key] = values.length === 1 ? values[0] : values;
  }

  return compacted;
}

export function compactSearchParams(path: string, params: URLSearchParams): URLSearchParams {
  const query: Record<string, string[]> = {};
  params.forEach((value, key) => {
    (query[key] ||= []).push(value);
  });

  const compacted = compactRouteQuery(path, query);
  const next = new URLSearchParams();
  for (const [key, value] of Object.entries(compacted)) {
    const values = Array.isArray(value) ? value : [value];
    values.forEach((item) => next.append(key, item));
  }
  return next;
}

export function compactInternalHref(href: unknown, fallback = "#"): string {
  const raw = String(href ?? "").trim();
  if (!raw) {
    return fallback;
  }
  if (raw.startsWith("#") || /^(?:mailto|tel|javascript):/i.test(raw)) {
    return raw;
  }

  const base = typeof window === "undefined" ? "http://localhost" : window.location.origin;
  let url: URL;
  try {
    url = new URL(raw, base);
  } catch {
    return raw;
  }
  if (url.origin !== base) {
    return raw;
  }

  const query = compactSearchParams(url.pathname, url.searchParams).toString();
  return `${url.pathname}${query ? `?${query}` : ""}${url.hash}`;
}
