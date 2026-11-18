import { useEffect, useState } from "react";
import type { Period } from "@/lib/statistics-period";
import type { PlayerStatistics } from "@/types/statistics";

type Result =
  | { key: string; status: "ok"; data: PlayerStatistics }
  | { key: string; status: "no-player" }
  | { key: string; status: "error"; message: string };

export interface PlayerStatisticsState {
  data: PlayerStatistics | null;
  isLoading: boolean;
  noPlayer: boolean;
  error: string | null;
}

/** Loads the current user's statistics for a club and period. */
export function usePlayerStatistics(
  clubId: string | null,
  period: Period | null,
): PlayerStatisticsState {
  const [result, setResult] = useState<Result | null>(null);

  const params = new URLSearchParams();
  if (period?.from) params.set("from", period.from);
  if (period?.to) params.set("to", period.to);
  const key = clubId && period ? `/api/clubs/${clubId}/players/me/statistics?${params}` : null;

  useEffect(() => {
    if (!key) return;
    const controller = new AbortController();
    fetch(key, { credentials: "include", signal: controller.signal })
      .then(async (r) => {
        if (r.status === 404) {
          setResult({ key, status: "no-player" });
          return;
        }
        if (!r.ok) {
          const body = await r.json().catch(() => null);
          throw new Error(body?.message ?? "Statistik konnte nicht geladen werden");
        }
        const data: PlayerStatistics = await r.json();
        setResult({ key, status: "ok", data });
      })
      .catch((e: unknown) => {
        if (controller.signal.aborted) return;
        setResult({ key, status: "error", message: e instanceof Error ? e.message : String(e) });
      });
    return () => controller.abort();
  }, [key]);

  // Keep showing the previous data while a new period loads (avoids layout jumps).
  const current = result && result.key === key ? result : null;
  return {
    data: result?.status === "ok" ? result.data : null,
    isLoading: key !== null && current === null,
    noPlayer: current?.status === "no-player",
    error: current?.status === "error" ? current.message : null,
  };
}
