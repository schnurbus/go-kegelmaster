import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

/** Error of an admin API request; status 404 also means "not an instance admin". */
export class AdminApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

type AdminFetchState<T> = {
  url: string;
  data: T | null;
  error: AdminApiError | null;
};

/**
 * Loads a read-only admin endpoint (GET, cookie session). Reloads whenever the URL
 * changes; `loading` is true until the response for the current URL arrived.
 */
export function useAdminFetch<T>(url: string) {
  const [state, setState] = useState<AdminFetchState<T> | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    (async () => {
      try {
        const resp = await fetch(url, { credentials: "include", signal: controller.signal });
        if (!resp.ok) {
          const body = await resp.json().catch(() => ({}));
          throw new AdminApiError(resp.status, body.message || `Fehler ${resp.status}`);
        }
        const data: T = await resp.json();
        setState({ url, data, error: null });
      } catch (err) {
        if (controller.signal.aborted) return;
        const error =
          err instanceof AdminApiError ? err : new AdminApiError(0, "Daten konnten nicht geladen werden");
        setState({ url, data: null, error });
      }
    })();
    return () => controller.abort();
  }, [url]);

  const current = state?.url === url ? state : null;
  return {
    data: current?.data ?? null,
    error: current?.error ?? null,
    loading: current === null,
  };
}

export function formatAdminDate(value: string | null | undefined): string {
  if (!value) return "–";
  return new Date(value).toLocaleDateString("de-DE", { dateStyle: "medium" });
}

export function formatAdminDateTime(value: string | null | undefined): string {
  if (!value) return "–";
  return new Date(value).toLocaleString("de-DE", { dateStyle: "medium", timeStyle: "short" });
}

export const MEMBERSHIPS_CSV_URL = "/api/admin/export/memberships.csv";

export type SortOrder = "asc" | "desc";

/**
 * List state (search, filter, sort, order, page) kept in the URL query, so lists
 * can be linked (e.g. /admin/users?filter=orphaned) and survive reloads.
 * `apiUrl` is the matching admin API URL.
 */
export function useAdminListParams(endpoint: string, defaultSort: string, defaultOrder: SortOrder) {
  const [params, setParams] = useSearchParams();
  const search = params.get("search") ?? "";
  const filter = params.get("filter") ?? "";
  const sort = params.get("sort") ?? defaultSort;
  const order: SortOrder = params.get("order") === "asc" ? "asc" : params.get("order") === "desc" ? "desc" : defaultOrder;
  const page = Math.max(1, Number(params.get("page")) || 1);

  const update = useCallback(
    (changes: Record<string, string | null>) => {
      setParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          for (const [key, value] of Object.entries(changes)) {
            if (value === null || value === "") next.delete(key);
            else next.set(key, value);
          }
          // Every change except paging starts again on page 1.
          if (!("page" in changes)) next.delete("page");
          return next;
        },
        { replace: true },
      );
    },
    [setParams],
  );

  const setSearch = useCallback((value: string) => update({ search: value.trim() || null }), [update]);
  const setFilter = useCallback((value: string) => update({ filter: value || null }), [update]);
  const setPage = useCallback(
    (value: number) => update({ page: value > 1 ? String(value) : null }),
    [update],
  );

  const query = new URLSearchParams();
  if (search) query.set("search", search);
  if (filter) query.set("filter", filter);
  query.set("sort", sort);
  query.set("order", order);
  query.set("page", String(page));
  query.set("limit", "25");

  return {
    search,
    filter,
    sort,
    order,
    page,
    apiUrl: `${endpoint}?${query.toString()}`,
    setSearch,
    setFilter,
    setPage,
    toggleSort: (field: string) =>
      update({
        sort: field,
        order: sort === field ? (order === "asc" ? "desc" : "asc") : field === "email" || field === "name" ? "asc" : "desc",
      }),
  };
}
