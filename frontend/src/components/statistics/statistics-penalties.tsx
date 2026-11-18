import { useMemo } from "react";
import { Bar, CartesianGrid, ComposedChart, Line, XAxis, YAxis } from "recharts";
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatDate, formatShortDate } from "@/lib/statistics-period";
import { formatCentsToEuro } from "@/types/player";
import type { PlayerStatistics } from "@/types/statistics";

const perGameDayConfig = {
  own: { label: "Meine Strafen", color: "var(--chart-1)" },
  club: { label: "Klub-Ø", color: "var(--chart-2)" },
} satisfies ChartConfig;

const qtyFormat = new Intl.NumberFormat("de-DE", { maximumFractionDigits: 2 });

export function PenaltiesPerGameDayChart({ stats }: { stats: PlayerStatistics }) {
  const data = useMemo(
    () =>
      stats.game_days.map((gd) => ({
        date: gd.date,
        own: gd.penalty_amount != null ? gd.penalty_amount / 100 : null,
        club: gd.club_average_amount / 100,
        attended: gd.attended,
      })),
    [stats.game_days],
  );

  return (
    <Card className="@container/card">
      <CardHeader>
        <CardTitle>Strafen pro Spieltag</CardTitle>
        <CardDescription>
          Eigene Strafen im Vergleich zum Durchschnitt aller Teilnehmer des Abends
        </CardDescription>
      </CardHeader>
      <CardContent className="px-2 pt-2 sm:px-6">
        {data.length === 0 ? (
          <p className="text-sm text-muted-foreground">Keine Spieltage im gewählten Zeitraum.</p>
        ) : (
          <ChartContainer config={perGameDayConfig} className="aspect-auto h-[260px] w-full">
            <ComposedChart data={data} margin={{ left: 4, right: 12 }}>
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                minTickGap={24}
                tickFormatter={(v: string) => formatDate(v, { day: "numeric", month: "short" })}
              />
              <YAxis
                tickLine={false}
                axisLine={false}
                width={48}
                tickFormatter={(v: number) => `${v} €`}
              />
              <ChartTooltip
                cursor={false}
                content={
                  <ChartTooltipContent
                    labelFormatter={(v) => formatDate(String(v))}
                    formatter={(value, name, item) => (
                      <div className="flex w-full items-center justify-between gap-4">
                        <span className="flex items-center gap-1.5 text-muted-foreground">
                          <span
                            className="size-2.5 rounded-[2px]"
                            style={{ background: item.color }}
                          />
                          {perGameDayConfig[name as keyof typeof perGameDayConfig]?.label ?? name}
                        </span>
                        <span className="font-mono font-medium tabular-nums">
                          {formatCentsToEuro(Math.round(Number(value) * 100))}
                        </span>
                      </div>
                    )}
                  />
                }
              />
              <ChartLegend content={<ChartLegendContent />} />
              <Bar dataKey="own" fill="var(--color-own)" radius={[4, 4, 0, 0]} maxBarSize={32} />
              <Line
                dataKey="club"
                type="monotone"
                stroke="var(--color-club)"
                strokeWidth={2}
                strokeDasharray="4 4"
                dot={false}
              />
            </ComposedChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}

export function PenaltiesByType({ stats }: { stats: PlayerStatistics }) {
  const rows = stats.penalties.by_type;
  const total = stats.penalties.total_amount;

  return (
    <Card className="@container/card">
      <CardHeader>
        <CardTitle>Strafen nach Typ</CardTitle>
        <CardDescription>Anzahl und Betrag je Strafe; Ø pro besuchtem Spieltag im Klubvergleich</CardDescription>
      </CardHeader>
      <CardContent>
        {rows.length === 0 ? (
          <p className="text-sm text-muted-foreground">Keine Strafen im gewählten Zeitraum. Vorbildlich!</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Strafe</TableHead>
                <TableHead className="text-right">Anzahl</TableHead>
                <TableHead className="text-right">Betrag</TableHead>
                <TableHead className="hidden text-right sm:table-cell">Ø / Abend</TableHead>
                <TableHead className="hidden text-right sm:table-cell">Klub-Ø</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((r) => {
                const share = total > 0 ? (r.amount / total) * 100 : 0;
                const above = r.average_per_game_day > r.club_average_per_game_day;
                return (
                  <TableRow key={r.penalty_type_id}>
                    <TableCell>
                      <div className="font-medium">{r.name}</div>
                      <div className="mt-1 h-1.5 w-full max-w-40 rounded-full bg-muted">
                        <div
                          className="h-1.5 rounded-full bg-[var(--chart-1)]"
                          style={{ width: `${Math.max(share, 2)}%` }}
                        />
                      </div>
                    </TableCell>
                    <TableCell className="text-right tabular-nums">{qtyFormat.format(r.quantity)}</TableCell>
                    <TableCell className="text-right tabular-nums">{formatCentsToEuro(r.amount)}</TableCell>
                    <TableCell
                      className={`hidden text-right tabular-nums sm:table-cell ${above ? "text-destructive" : ""}`}
                    >
                      {qtyFormat.format(r.average_per_game_day)}
                    </TableCell>
                    <TableCell className="hidden text-right tabular-nums text-muted-foreground sm:table-cell">
                      {qtyFormat.format(r.club_average_per_game_day)}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}

export function Highlights({ stats }: { stats: PlayerStatistics }) {
  const { attendance, penalties, competitions } = stats;
  const topType = penalties.by_type[0];
  const mostWins = [...competitions].sort((a, b) => b.wins - a.wins)[0];

  const items: Array<{ label: string; value: string; hint?: string }> = [];
  if (penalties.most_expensive_game_day) {
    items.push({
      label: "Teuerster Abend",
      value: formatCentsToEuro(penalties.most_expensive_game_day.amount),
      hint: formatDate(penalties.most_expensive_game_day.date),
    });
  }
  if (penalties.cheapest_game_day) {
    items.push({
      label: "Günstigster Abend",
      value: formatCentsToEuro(penalties.cheapest_game_day.amount),
      hint: formatDate(penalties.cheapest_game_day.date),
    });
  }
  if (topType) {
    items.push({
      label: "Teuerste Strafe",
      value: topType.name,
      hint: `${qtyFormat.format(topType.quantity)}× · ${formatCentsToEuro(topType.amount)}`,
    });
  }
  if (mostWins && mostWins.wins > 0) {
    items.push({
      label: "Erfolgreichster Wettbewerb",
      value: mostWins.name,
      hint: `${mostWins.wins} ${mostWins.wins === 1 ? "Sieg" : "Siege"}`,
    });
  }
  items.push({
    label: "Längste Teilnahme-Serie",
    value: `${attendance.longest_streak} ${attendance.longest_streak === 1 ? "Spieltag" : "Spieltage"} in Folge`,
  });
  if (attendance.first_attended && attendance.last_attended) {
    items.push({
      label: "Erster / letzter Besuch",
      value: `${formatShortDate(attendance.first_attended)} – ${formatShortDate(attendance.last_attended)}`,
    });
  }

  return (
    <Card className="@container/card">
      <CardHeader>
        <CardTitle>Highlights</CardTitle>
        <CardDescription>Bemerkenswertes aus dem gewählten Zeitraum</CardDescription>
      </CardHeader>
      <CardContent>
        <dl className="grid gap-4 @md/card:grid-cols-2">
          {items.map((it) => (
            <div key={it.label} className="flex flex-col gap-0.5">
              <dt className="text-xs text-muted-foreground">{it.label}</dt>
              <dd className="font-semibold">{it.value}</dd>
              {it.hint && <dd className="text-xs text-muted-foreground">{it.hint}</dd>}
            </div>
          ))}
        </dl>
      </CardContent>
    </Card>
  );
}
