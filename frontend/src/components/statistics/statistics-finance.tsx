import { Area, AreaChart, CartesianGrid, ReferenceLine, XAxis, YAxis } from "recharts";
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
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { formatDate } from "@/lib/statistics-period";
import { formatCentsToEuro } from "@/types/player";
import type { PlayerStatistics } from "@/types/statistics";

const chartConfig = {
  balance: { label: "Kontostand", color: "var(--chart-2)" },
} satisfies ChartConfig;

export function BalanceChart({ stats }: { stats: PlayerStatistics }) {
  const data = stats.finance.balance_history.map((p) => ({ date: p.date, balance: p.balance / 100 }));

  return (
    <Card className="@container/card">
      <CardHeader>
        <CardTitle>Kontoverlauf</CardTitle>
        <CardDescription>
          {formatCentsToEuro(stats.finance.start_balance)} zu Beginn →{" "}
          {formatCentsToEuro(stats.finance.end_balance)} am Ende des Zeitraums
        </CardDescription>
      </CardHeader>
      <CardContent className="px-2 pt-2 sm:px-6">
        {data.length < 2 ? (
          <p className="text-sm text-muted-foreground">Keine Buchungen im gewählten Zeitraum.</p>
        ) : (
          <ChartContainer config={chartConfig} className="aspect-auto h-[260px] w-full">
            <AreaChart data={data} margin={{ left: 4, right: 12 }}>
              <defs>
                <linearGradient id="fillBalance" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--color-balance)" stopOpacity={0.5} />
                  <stop offset="95%" stopColor="var(--color-balance)" stopOpacity={0.05} />
                </linearGradient>
              </defs>
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                minTickGap={24}
                tickFormatter={(v: string) => formatDate(v, { day: "numeric", month: "short" })}
              />
              <YAxis tickLine={false} axisLine={false} width={56} tickFormatter={(v: number) => `${v} €`} />
              <ReferenceLine y={0} stroke="var(--border)" />
              <ChartTooltip
                cursor={false}
                content={
                  <ChartTooltipContent
                    labelFormatter={(v) => formatDate(String(v))}
                    formatter={(value) => (
                      <span className="font-mono font-medium tabular-nums">
                        {formatCentsToEuro(Math.round(Number(value) * 100))}
                      </span>
                    )}
                  />
                }
              />
              <Area
                dataKey="balance"
                type="stepAfter"
                stroke="var(--color-balance)"
                fill="url(#fillBalance)"
                strokeWidth={2}
              />
            </AreaChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
