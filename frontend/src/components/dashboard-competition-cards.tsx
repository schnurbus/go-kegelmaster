"use client";

import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useClub } from "@/context/ClubContext";
import {
  type DashboardCompetitionData,
  type GameDayDetailParticipant,
} from "@/hooks/use-dashboard-competition-data";

export function DashboardLastGameDayResults({
  data,
}: {
  data: DashboardCompetitionData;
}) {
  const { activeClub } = useClub();
  const {
    competitions,
    lastGameDayDetail,
    myPlayer,
    isLoading,
    error,
  } = data;

  if (!activeClub || competitions.length === 0) return null;
  if (error) return null;

  const compById = new Map(competitions.map((c) => [c.id, c]));
  const myParticipant = myPlayer && lastGameDayDetail
    ? lastGameDayDetail.participants.find(
        (p) => p.participant.player_id === myPlayer.id
      )
    : undefined;
  const myValues = myParticipant?.competition_values ?? [];

  const formatGameDayDate = (dateStr: string) => {
    const d = new Date(dateStr + "T12:00:00");
    return d.toLocaleDateString("de-DE", {
      day: "numeric",
      month: "short",
      year: "numeric",
    });
  };

  const statusLine =
    isLoading
      ? "Laden..."
      : !lastGameDayDetail
        ? "Kein Spieltag vorhanden"
        : !myParticipant
          ? "Am letzten Spieltag nicht teilgenommen"
          : formatGameDayDate(lastGameDayDetail.game_day.date);

  return (
    <Card className="@container/card">
      <CardHeader className="relative pb-1">
        <CardDescription>Meine Wettbewerbsergebnisse</CardDescription>
        <CardTitle className="@[250px]/card:text-2xl text-xl font-semibold tabular-nums">
          {statusLine}
        </CardTitle>
      </CardHeader>
      <CardFooter className="flex-col items-start gap-0.5 pt-1 text-sm">
        {!isLoading && lastGameDayDetail && myParticipant && (
          myValues.length === 0 ? (
            <span className="text-muted-foreground">
              Keine Wettbewerbswerte
            </span>
          ) : (
            <span className="text-muted-foreground">
              {myValues
                .map((cv) => {
                  const comp = compById.get(cv.competition_id);
                  return `${comp?.name ?? cv.competition_id}: ${cv.value}`;
                })
                .join(" · ")}
            </span>
          )
        )}
      </CardFooter>
    </Card>
  );
}

function getWinnersLosers(
  participants: GameDayDetailParticipant[],
  competitionId: string,
  scoringType: "winner" | "loser" | "both",
  isGenderSpecific: boolean
): {
  winnerMale: string | null;
  winnerFemale: string | null;
  loserMale: string | null;
  loserFemale: string | null;
  winner: string | null;
  loser: string | null;
} {
  const entries: Array<{
    playerName: string;
    gender: string | null;
    value: number;
  }> = [];
  for (const p of participants) {
    const cv = p.competition_values.find((c) => c.competition_id === competitionId);
    if (cv == null) continue;
    const gender = p.participant.player_gender ?? null;
    entries.push({
      playerName: p.participant.player_name,
      gender,
      value: cv.value,
    });
  }

  const withGender = (g: string | null) =>
    entries.filter((e) => e.gender === g && e.gender != null);

  let winnerMale: string | null = null;
  let winnerFemale: string | null = null;
  let loserMale: string | null = null;
  let loserFemale: string | null = null;
  let winner: string | null = null;
  let loser: string | null = null;

  if (isGenderSpecific) {
    for (const gender of ["male", "female"] as const) {
      const list = withGender(gender);
      if (list.length === 0) continue;
      const maxVal = Math.max(...list.map((e) => e.value));
      const minVal = Math.min(...list.map((e) => e.value));
      const maxEntry = list.find((e) => e.value === maxVal);
      const minEntry = list.find((e) => e.value === minVal);
      if (gender === "male") {
        winnerMale = (scoringType === "winner" || scoringType === "both") && maxEntry ? maxEntry.playerName : null;
        loserMale = (scoringType === "loser" || scoringType === "both") && minEntry ? minEntry.playerName : null;
      } else {
        winnerFemale = (scoringType === "winner" || scoringType === "both") && maxEntry ? maxEntry.playerName : null;
        loserFemale = (scoringType === "loser" || scoringType === "both") && minEntry ? minEntry.playerName : null;
      }
    }
  } else {
    if (entries.length === 0) {
      // no participants with value
    } else {
      const maxVal = Math.max(...entries.map((e) => e.value));
      const minVal = Math.min(...entries.map((e) => e.value));
      const maxEntry = entries.find((e) => e.value === maxVal);
      const minEntry = entries.find((e) => e.value === minVal);
      winner = (scoringType === "winner" || scoringType === "both") && maxEntry ? maxEntry.playerName : null;
      loser = (scoringType === "loser" || scoringType === "both") && minEntry ? minEntry.playerName : null;
    }
  }

  return {
    winnerMale,
    winnerFemale,
    loserMale,
    loserFemale,
    winner,
    loser,
  };
}

export function DashboardWinnersLosers({
  data,
}: {
  data: DashboardCompetitionData;
}) {
  const { activeClub } = useClub();
  const {
    competitions,
    lastGameDayDetail,
    isLoading,
    error,
  } = data;

  if (!activeClub || competitions.length === 0) return null;
  if (error || !lastGameDayDetail) return null;
  if (isLoading) return null;

  return (
    <>
      {competitions.map((comp) => {
        const { winnerMale, winnerFemale, loserMale, loserFemale, winner, loser } =
          getWinnersLosers(
            lastGameDayDetail.participants,
            comp.id,
            comp.scoring_type,
            comp.is_gender_specific
          );
        const hasAny =
          winner != null ||
          loser != null ||
          winnerMale != null ||
          winnerFemale != null ||
          loserMale != null ||
          loserFemale != null;
        if (!hasAny) return null;

        return (
          <Card key={comp.id} className="@container/card">
            <CardHeader className="relative">
              <CardDescription>Letzter Spieltag</CardDescription>
              <CardTitle className="@[250px]/card:text-lg text-base font-semibold">
                {comp.name}
              </CardTitle>
            </CardHeader>
            <CardFooter className="flex-col items-start gap-0.5 pt-1 text-sm">
              {comp.is_gender_specific ? (
                <>
                  {(winnerMale != null || loserMale != null) && (
                    <div className="text-muted-foreground">
                      Sieger / Verlierer:{" "}
                      <span className="font-medium text-foreground">
                        {[
                          (comp.scoring_type === "winner" || comp.scoring_type === "both") && winnerMale != null
                            ? winnerMale
                            : null,
                          (comp.scoring_type === "loser" || comp.scoring_type === "both") && loserMale != null
                            ? loserMale
                            : null,
                        ]
                          .filter(Boolean)
                          .join(" / ") || "—"}
                      </span>
                    </div>
                  )}
                  {(winnerFemale != null || loserFemale != null) && (
                    <div className="text-muted-foreground">
                      Sieger / Verlierer:{" "}
                      <span className="font-medium text-foreground">
                        {[
                          (comp.scoring_type === "winner" || comp.scoring_type === "both") && winnerFemale != null
                            ? winnerFemale
                            : null,
                          (comp.scoring_type === "loser" || comp.scoring_type === "both") && loserFemale != null
                            ? loserFemale
                            : null,
                        ]
                          .filter(Boolean)
                          .join(" / ") || "—"}
                      </span>
                    </div>
                  )}
                </>
              ) : (
                (winner != null || loser != null) && (
                  <div className="text-muted-foreground">
                    Sieger / Verlierer:{" "}
                    <span className="font-medium text-foreground">
                      {[winner, loser].filter(Boolean).join(" / ") || "—"}
                    </span>
                  </div>
                )
              )}
            </CardFooter>
          </Card>
        );
      })}
    </>
  );
}
