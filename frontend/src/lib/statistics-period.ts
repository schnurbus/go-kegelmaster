// Period presets for the statistics page. Dates are local calendar dates (YYYY-MM-DD).

export type PeriodPreset =
  | "3m"
  | "6m"
  | "12m"
  | "this-year"
  | "last-year"
  | "all"
  | "custom";

export const PERIOD_PRESETS: Array<{ value: PeriodPreset; label: string }> = [
  { value: "3m", label: "Letzte 3 Monate" },
  { value: "6m", label: "Letzte 6 Monate" },
  { value: "12m", label: "Letzte 12 Monate" },
  { value: "this-year", label: "Dieses Jahr" },
  { value: "last-year", label: "Letztes Jahr" },
  { value: "all", label: "Gesamter Zeitraum" },
  { value: "custom", label: "Benutzerdefiniert" },
];

export const DEFAULT_PRESET: Exclude<PeriodPreset, "custom"> = "12m";

export interface Period {
  from: string | null;
  to: string;
}

export function toISODate(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

export function isPreset(v: string | null): v is PeriodPreset {
  return PERIOD_PRESETS.some((p) => p.value === v);
}

/** Resolves a preset to a concrete period relative to `today`. */
export function resolvePreset(preset: Exclude<PeriodPreset, "custom">, today = new Date()): Period {
  const to = toISODate(today);
  const monthsAgo = (n: number) => {
    const d = new Date(today);
    d.setMonth(d.getMonth() - n);
    return toISODate(d);
  };
  switch (preset) {
    case "3m":
      return { from: monthsAgo(3), to };
    case "6m":
      return { from: monthsAgo(6), to };
    case "12m":
      return { from: monthsAgo(12), to };
    case "this-year":
      return { from: `${today.getFullYear()}-01-01`, to };
    case "last-year": {
      const y = today.getFullYear() - 1;
      return { from: `${y}-01-01`, to: `${y}-12-31` };
    }
    case "all":
      return { from: null, to };
  }
}

/** Formats an ISO date (YYYY-MM-DD) for display, e.g. "5. März 2026". */
export function formatDate(iso: string, opts?: Intl.DateTimeFormatOptions): string {
  return new Date(iso + "T12:00:00").toLocaleDateString(
    "de-DE",
    opts ?? { day: "numeric", month: "short", year: "numeric" },
  );
}

export function formatShortDate(iso: string): string {
  return formatDate(iso, { day: "2-digit", month: "2-digit", year: "2-digit" });
}
