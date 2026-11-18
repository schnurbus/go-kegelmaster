import { useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import { Loader2Icon } from "lucide-react";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PeriodPicker } from "@/components/statistics/period-picker";
import { StatisticsKpis } from "@/components/statistics/statistics-kpis";
import {
  Highlights,
  PenaltiesByType,
  PenaltiesPerGameDayChart,
} from "@/components/statistics/statistics-penalties";
import { StatisticsCompetitions } from "@/components/statistics/statistics-competitions";
import { BalanceChart } from "@/components/statistics/statistics-finance";
import { useClub } from "@/context/ClubContext";
import { usePlayerStatistics } from "@/hooks/use-player-statistics";
import {
  DEFAULT_PRESET,
  formatDate,
  isPreset,
  type Period,
  type PeriodPreset,
  resolvePreset,
  toISODate,
} from "@/lib/statistics-period";

const ISO_DATE = /^\d{4}-\d{2}-\d{2}$/;

function Message({ title, text }: { title: string; text: string }) {
  return (
    <div className="px-4 lg:px-6">
      <Card>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
          <CardDescription>{text}</CardDescription>
        </CardHeader>
      </Card>
    </div>
  );
}

function StatisticsPage() {
  const { activeClub } = useClub();
  const [searchParams, setSearchParams] = useSearchParams();

  // Period lives in the URL (?range=…&from=…&to=…) so it survives reloads and can be shared.
  const rangeParam = searchParams.get("range");
  const preset: PeriodPreset = isPreset(rangeParam) ? rangeParam : DEFAULT_PRESET;
  const customFrom = searchParams.get("from") ?? "";
  const customTo = searchParams.get("to") ?? "";

  const period: Period | null = useMemo(() => {
    if (preset !== "custom") return resolvePreset(preset);
    if (!ISO_DATE.test(customFrom) || !ISO_DATE.test(customTo) || customFrom > customTo) return null;
    return { from: customFrom, to: customTo };
  }, [preset, customFrom, customTo]);

  const { data, isLoading, noPlayer, error } = usePlayerStatistics(activeClub?.id ?? null, period);

  const setPreset = (next: PeriodPreset) => {
    if (next === "custom") {
      // Start the custom range from the currently shown period.
      const base = period ?? resolvePreset(DEFAULT_PRESET);
      const from = base.from ?? data?.attendance.first_attended ?? toISODate(new Date(new Date().getFullYear(), 0, 1));
      setSearchParams({ range: "custom", from, to: base.to }, { replace: true });
    } else {
      setSearchParams(next === DEFAULT_PRESET ? {} : { range: next }, { replace: true });
    }
  };

  const setCustom = (from: string, to: string) =>
    setSearchParams({ range: "custom", from, to }, { replace: true });

  if (!activeClub) {
    return (
      <div className="py-4 md:py-6">
        <Message title="Statistiken" text="Kein Club ausgewählt. Bitte wählen Sie einen Club." />
      </div>
    );
  }

  const periodLabel = period
    ? period.from
      ? `${formatDate(period.from)} – ${formatDate(period.to)}`
      : `Alle Spieltage bis ${formatDate(period.to)}`
    : "Bitte einen gültigen Zeitraum wählen";

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="flex flex-wrap items-end justify-between gap-3 px-4 lg:px-6">
        <div className="min-w-0">
          <h1 className="text-2xl font-semibold tracking-tight">
            {data ? `Statistiken von ${data.player.name}` : "Meine Statistiken"}
          </h1>
          <p className="flex items-center gap-2 text-sm text-muted-foreground">
            {periodLabel}
            {isLoading && <Loader2Icon className="size-3.5 animate-spin" aria-label="Laden" />}
          </p>
        </div>
        <PeriodPicker
          preset={preset}
          customFrom={customFrom}
          customTo={customTo}
          onPresetChange={setPreset}
          onCustomChange={setCustom}
        />
      </div>

      {noPlayer ? (
        <Message
          title="Kein Spieler in diesem Club"
          text="Ihrem Benutzerkonto ist in diesem Club kein Spieler zugeordnet. Bitten Sie den Club-Verwalter um eine Einladung."
        />
      ) : error ? (
        <Message title="Fehler" text={error} />
      ) : !data ? (
        isLoading ? (
          <Message title="Statistiken" text="Laden..." />
        ) : null
      ) : (
        <div className={`flex flex-col gap-4 md:gap-6 ${isLoading ? "opacity-60 transition-opacity" : ""}`}>
          <StatisticsKpis stats={data} />
          <div className="px-4 lg:px-6">
            <PenaltiesPerGameDayChart stats={data} />
          </div>
          <div className="grid grid-cols-1 gap-4 px-4 @5xl/main:grid-cols-5 lg:px-6">
            <div className="@5xl/main:col-span-3">
              <PenaltiesByType stats={data} />
            </div>
            <div className="@5xl/main:col-span-2">
              <Highlights stats={data} />
            </div>
          </div>
          <StatisticsCompetitions competitions={data.competitions} />
          <div className="px-4 lg:px-6">
            <BalanceChart stats={data} />
          </div>
          <p className="px-4 text-xs text-muted-foreground lg:px-6">
            Es werden nur veröffentlichte Spieltage berücksichtigt. Klub-Durchschnitte beziehen sich auf alle
            Teilnahmen im Zeitraum.
          </p>
        </div>
      )}
    </div>
  );
}

export default StatisticsPage;
