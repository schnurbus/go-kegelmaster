import { CartesianGrid, Line, LineChart, XAxis, YAxis } from "recharts";
import { TrophyIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
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
import type { CompetitionStatistics } from "@/types/statistics";

const chartConfig = {
  value: { label: "Mein Wert", color: "var(--chart-1)" },
  club_average: { label: "Klub-Ø", color: "var(--chart-2)" },
} satisfies ChartConfig;

const numFormat = new Intl.NumberFormat("de-DE", { maximumFractionDigits: 2 });

function Stat({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex flex-col">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="font-semibold tabular-nums">{value}</span>
    </div>
  );
}

function CompetitionCard({ comp }: { comp: CompetitionStatistics }) {
  const showWins = comp.scoring_type !== "loser";
  const showLosses = comp.scoring_type !== "winner";

  return (
    <Card className="@container/card">
      <CardHeader className="relative">
        <CardDescription>
          Wettbewerb{comp.is_gender_specific ? " · getrennt nach Geschlecht" : ""}
        </CardDescription>
        <CardTitle className="text-lg">{comp.name}</CardTitle>
        {comp.rank > 0 && (
          <div className="absolute right-4 top-0">
            <Badge variant="outline">
              {comp.rank === 1 && <TrophyIcon className="size-3 text-amber-500" />}
              Platz {comp.rank} / {comp.rank_of}
            </Badge>
          </div>
        )}
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="grid grid-cols-3 gap-3 @md/card:grid-cols-6">
          <Stat label="Teilnahmen" value={comp.participations} />
          <Stat label="Ø-Wert" value={numFormat.format(comp.average)} />
          <Stat label="Klub-Ø" value={numFormat.format(comp.club_average)} />
          <Stat label="Bestwert" value={comp.best ?? "–"} />
          {showWins && <Stat label="Siege" value={comp.wins} />}
          {showLosses && <Stat label="Letzter Platz" value={comp.losses} />}
          {!(showWins && showLosses) && <Stat label="Schlechtester" value={comp.worst ?? "–"} />}
        </div>
        {comp.history.length > 1 && (
          <ChartContainer config={chartConfig} className="aspect-auto h-[160px] w-full">
            <LineChart data={comp.history} margin={{ left: 0, right: 8, top: 4 }}>
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                minTickGap={24}
                tickFormatter={(v: string) => formatDate(v, { day: "numeric", month: "short" })}
              />
              <YAxis tickLine={false} axisLine={false} width={32} allowDecimals={false} />
              <ChartTooltip
                cursor={false}
                content={<ChartTooltipContent labelFormatter={(v) => formatDate(String(v))} indicator="dot" />}
              />
              <Line
                dataKey="value"
                type="monotone"
                stroke="var(--color-value)"
                strokeWidth={2}
                dot={(props) => {
                  const { cx, cy, payload, index } = props as {
                    cx: number;
                    cy: number;
                    index: number;
                    payload: CompetitionStatistics["history"][number];
                  };
                  const fill = payload.is_win
                    ? "var(--color-chart-4)"
                    : payload.is_loss
                      ? "var(--destructive)"
                      : "var(--color-value)";
                  const r = payload.is_win || payload.is_loss ? 4 : 2.5;
                  return <circle key={index} cx={cx} cy={cy} r={r} fill={fill} />;
                }}
              />
              <Line
                dataKey="club_average"
                type="monotone"
                stroke="var(--color-club_average)"
                strokeWidth={1.5}
                strokeDasharray="4 4"
                dot={false}
              />
            </LineChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}

export function StatisticsCompetitions({ competitions }: { competitions: CompetitionStatistics[] }) {
  const relevant = competitions.filter((c) => c.participations > 0);
  if (relevant.length === 0) return null;

  return (
    <section className="flex flex-col gap-3 px-4 lg:px-6">
      <div>
        <h2 className="text-lg font-semibold">Wettbewerbe</h2>
        <p className="text-sm text-muted-foreground">
          Siege und letzte Plätze pro Abend (größere Punkte im Verlauf), Platz nach Ø-Wert im Klub
        </p>
      </div>
      <div className="grid grid-cols-1 gap-4 @3xl/main:grid-cols-2">
        {relevant.map((c) => (
          <CompetitionCard key={c.competition_id} comp={c} />
        ))}
      </div>
    </section>
  );
}
