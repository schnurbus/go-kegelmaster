import type { ReactNode } from "react";
import { TrendingDownIcon, TrendingUpIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { formatCentsToEuro } from "@/types/player";
import type { PlayerStatistics } from "@/types/statistics";

export function KpiCard({
  label,
  value,
  badge,
  children,
}: {
  label: string;
  value: ReactNode;
  badge?: ReactNode;
  children?: ReactNode;
}) {
  return (
    <Card className="@container/card">
      <CardHeader className="relative">
        <CardDescription>{label}</CardDescription>
        <CardTitle className="@[250px]/card:text-3xl text-2xl font-semibold tabular-nums">
          {value}
        </CardTitle>
        {badge && <div className="absolute right-4 top-0">{badge}</div>}
      </CardHeader>
      {children && (
        <CardFooter className="flex-col items-start gap-1 text-sm text-muted-foreground">
          {children}
        </CardFooter>
      )}
    </Card>
  );
}

/** Compares own value to the club average; for penalties lower is "better". */
function ComparisonBadge({ own, club }: { own: number; club: number }) {
  if (club <= 0) return null;
  const diff = Math.round(((own - club) / club) * 100);
  if (diff === 0) return <Badge variant="outline">= Klub-Ø</Badge>;
  const Icon = diff > 0 ? TrendingUpIcon : TrendingDownIcon;
  return (
    <Badge variant="outline" className={diff > 0 ? "text-destructive" : "text-emerald-600 dark:text-emerald-400"}>
      <Icon className="size-3" />
      {diff > 0 ? "+" : ""}
      {diff} %
    </Badge>
  );
}

export function StatisticsKpis({ stats }: { stats: PlayerStatistics }) {
  const { attendance, penalties, finance } = stats;
  const ratePct = Math.round(attendance.rate * 100);
  const balanceDelta = finance.end_balance - finance.start_balance;

  return (
    <div className="*:data-[slot=card]:shadow-xs @xl/main:grid-cols-2 @5xl/main:grid-cols-4 grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:bg-gradient-to-t *:data-[slot=card]:from-primary/5 *:data-[slot=card]:to-card dark:*:data-[slot=card]:bg-card lg:px-6">
      <KpiCard
        label="Teilnahme"
        value={`${attendance.attended} / ${attendance.club_game_days}`}
        badge={attendance.club_game_days > 0 && <Badge variant="outline">{ratePct} %</Badge>}
      >
        <div className="font-medium text-foreground">
          {attendance.club_game_days === 0
            ? "Keine Spieltage im Zeitraum"
            : `${ratePct} % der Spieltage besucht`}
        </div>
        <div>
          Längste Serie: {attendance.longest_streak} · Aktuell: {attendance.current_streak}
        </div>
      </KpiCard>

      <KpiCard
        label="Strafen gesamt"
        value={formatCentsToEuro(penalties.total_amount)}
        badge={
          attendance.attended > 0 && (
            <ComparisonBadge own={penalties.average_per_game_day} club={penalties.club_average_per_game_day} />
          )
        }
      >
        <div className="font-medium text-foreground">
          Ø {formatCentsToEuro(penalties.average_per_game_day)} pro Spieltag
        </div>
        <div>Klub-Ø: {formatCentsToEuro(penalties.club_average_per_game_day)} pro Spieltag</div>
      </KpiCard>

      <KpiCard
        label="Platz im Klub (Ø-Strafen)"
        value={penalties.rank > 0 ? `${penalties.rank}. von ${penalties.rank_of}` : "–"}
      >
        <div className="font-medium text-foreground">
          {penalties.rank === 0
            ? "Keine Teilnahme im Zeitraum"
            : penalties.rank === 1
              ? "Spitzenreiter der Klubkasse"
              : penalties.rank === penalties.rank_of
                ? "Sparsamster Kegler im Klub"
                : "Platz 1 = höchste Ø-Strafen"}
        </div>
        <div>
          {penalties.penalty_free_game_days} strafenfreie{" "}
          {penalties.penalty_free_game_days === 1 ? "Spieltag" : "Spieltage"}
        </div>
      </KpiCard>

      <KpiCard
        label="Kontostand"
        value={formatCentsToEuro(finance.end_balance)}
        badge={
          <Badge
            variant="outline"
            className={balanceDelta < 0 ? "text-destructive" : balanceDelta > 0 ? "text-emerald-600 dark:text-emerald-400" : ""}
          >
            {balanceDelta > 0 ? "+" : ""}
            {formatCentsToEuro(balanceDelta)}
          </Badge>
        }
      >
        <div className="font-medium text-foreground">
          Einzahlungen: {formatCentsToEuro(finance.deposits)}
        </div>
        <div>
          Strafen {formatCentsToEuro(finance.fees)} · Grundgebühr {formatCentsToEuro(finance.base_fees)}
        </div>
      </KpiCard>
    </div>
  );
}
