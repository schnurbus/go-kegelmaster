"use client";

import { useEffect, useMemo, useState } from "react";
import { CartesianGrid, Line, LineChart, XAxis, YAxis } from "recharts";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  type ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { useClub } from "@/context/ClubContext";

interface PenaltyHistoryDay {
  game_day_id: string;
  date: string;
  penalties: Array<{
    penalty_type_id: string;
    penalty_type_name: string;
    quantity: number;
  }>;
}

// Use CSS variables directly (they resolve to oklch() in theme)
const CHART_COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

function sanitizeKey(name: string): string {
  return name.replace(/\s+/g, "_").replace(/[^a-zA-Z0-9_]/g, "") || "penalty";
}

export function DashboardPenaltyChart() {
  const { activeClub } = useClub();
  const [history, setHistory] = useState<PenaltyHistoryDay[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!activeClub) return;

    let cancelled = false;
    queueMicrotask(() => {
      if (!cancelled) {
        setError(null);
        setIsLoading(true);
      }
    });

    const since = new Date();
    since.setFullYear(since.getFullYear() - 1);
    const sinceStr = since.toISOString().slice(0, 10);

    fetch(
      `/api/clubs/${activeClub.id}/players/me/penalty-history?since=${sinceStr}`,
      { credentials: "include" }
    )
      .then((r) => {
        if (cancelled) return null;
        if (!r.ok) throw new Error("Fehler beim Laden");
        return r.json();
      })
      .then((data: PenaltyHistoryDay[] | null) => {
        if (!cancelled && data) setHistory(data);
      })
      .catch((e) => {
        if (!cancelled) {
          setError(String(e));
          setHistory([]);
        }
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [activeClub]);

  const { chartData, chartConfig, dataKeys } = useMemo(() => {
    const penaltyTypes = new Map<string, string>();
    for (const day of history) {
      for (const p of day.penalties) {
        if (!penaltyTypes.has(p.penalty_type_id)) {
          penaltyTypes.set(p.penalty_type_id, p.penalty_type_name);
        }
      }
    }
    const keys = Array.from(penaltyTypes.keys());
    const config: ChartConfig = {};
    keys.forEach((id, i) => {
      const name = penaltyTypes.get(id) ?? id;
      const key = sanitizeKey(name) + "_" + id.slice(0, 8);
      config[key] = {
        label: name,
        color: CHART_COLORS[i % CHART_COLORS.length],
      };
    });

    const byDate = new Map<string, Record<string, number>>();
    for (const day of history) {
      const row: Record<string, string | number> = { date: day.date };
      for (const id of keys) {
        const p = day.penalties.find((x) => x.penalty_type_id === id);
        const key = sanitizeKey(penaltyTypes.get(id) ?? id) + "_" + id.slice(0, 8);
        row[key] = p ? p.quantity : 0;
      }
      byDate.set(day.date, row as Record<string, number>);
    }

    const data = Array.from(byDate.entries())
      .map(([date, row]) => ({ ...row, date }))
      .sort((a, b) => a.date.localeCompare(b.date));

    const dataKeys = keys.map(
      (id) => sanitizeKey(penaltyTypes.get(id) ?? id) + "_" + id.slice(0, 8)
    );

    return { chartData: data, chartConfig: config, dataKeys };
  }, [history]);

  if (!activeClub) {
    return (
      <div className="px-4 lg:px-6">
        <Card className="@container/card">
          <CardHeader>
            <CardTitle>Strafen pro Spieltag</CardTitle>
            <CardDescription>
              Kein Club ausgewählt. Bitte wählen Sie einen Club.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="px-4 lg:px-6">
        <Card className="@container/card">
          <CardHeader>
            <CardTitle>Strafen pro Spieltag</CardTitle>
            <CardDescription>Laden...</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (error) {
    return (
      <div className="px-4 lg:px-6">
        <Card className="@container/card">
          <CardHeader>
            <CardTitle>Strafen pro Spieltag</CardTitle>
            <CardDescription className="text-destructive">{error}</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (chartData.length === 0) {
    return (
      <div className="px-4 lg:px-6">
        <Card className="@container/card">
          <CardHeader>
            <CardTitle>Strafen pro Spieltag</CardTitle>
            <CardDescription>
              Keine Daten der letzten 12 Monate. Sobald Sie an Spieltagen teilnehmen, erscheinen hier die Strafen pro Strafentyp.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="px-4 lg:px-6">
      <Card className="@container/card">
        <CardHeader className="relative">
          <CardTitle>Strafen pro Spieltag</CardTitle>
          <CardDescription>
            Anzahl der Strafen pro Strafentyp der letzten 12 Monate (ein Jahr).
          </CardDescription>
        </CardHeader>
        <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
          <ChartContainer
            config={chartConfig}
            className="aspect-auto h-[250px] w-full"
          >
            <LineChart data={chartData} margin={{ left: 12, right: 12 }}>
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                minTickGap={32}
                tickFormatter={(value) => {
                  const d = new Date(value);
                  return d.toLocaleDateString("de-DE", {
                    month: "short",
                    day: "numeric",
                  });
                }}
              />
              <YAxis
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                allowDecimals={false}
              />
              <ChartTooltip
                cursor={false}
                content={
                  <ChartTooltipContent
                    labelFormatter={(value) =>
                      new Date(value).toLocaleDateString("de-DE", {
                        month: "short",
                        day: "numeric",
                        year: "numeric",
                      })
                    }
                    indicator="dot"
                  />
                }
              />
              <ChartLegend
                verticalAlign="bottom"
                content={<ChartLegendContent />}
              />
              {dataKeys.map((key, i) => (
                <Line
                  key={key}
                  type="monotone"
                  name={String(chartConfig[key]?.label ?? key)}
                  dataKey={key}
                  stroke={chartConfig[key]?.color ?? CHART_COLORS[i % CHART_COLORS.length]}
                  strokeWidth={2}
                  dot={{ r: 2 }}
                  connectNulls
                />
              ))}
            </LineChart>
          </ChartContainer>
        </CardContent>
      </Card>
    </div>
  );
}
