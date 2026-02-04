"use client";

import { useEffect, useState } from "react";
import { useClub } from "@/context/ClubContext";
import type { Competition } from "@/types/competition";
import type { Player } from "@/types/player";

export interface GameDayDetailParticipant {
  participant: {
    id: string;
    game_day_id: string;
    player_id: string;
    player_name: string;
    player_gender?: string | null;
    created_at: string;
  };
  fees: unknown[];
  competition_values: Array<{
    id: string;
    game_day_participant_id: string;
    competition_id: string;
    value: number;
    created_at: string;
    updated_at: string;
  }>;
}

export interface GameDayDetail {
  game_day: {
    id: string;
    club_id: string;
    date: string;
    notes: string;
    created_at: string;
    updated_at: string;
  };
  participants: GameDayDetailParticipant[];
}

export interface DashboardCompetitionData {
  competitions: Competition[];
  lastGameDayDetail: GameDayDetail | null;
  myPlayer: Player | null;
  isLoading: boolean;
  error: string | null;
}

export function useDashboardCompetitionData(): DashboardCompetitionData {
  const { activeClub } = useClub();
  const [competitions, setCompetitions] = useState<Competition[]>([]);
  const [lastGameDayDetail, setLastGameDayDetail] =
    useState<GameDayDetail | null>(null);
  const [myPlayer, setMyPlayer] = useState<Player | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!activeClub) {
      setCompetitions([]);
      setLastGameDayDetail(null);
      setMyPlayer(null);
      setIsLoading(false);
      return;
    }

    let cancelled = false;
    setError(null);
    setIsLoading(true);

    const run = async () => {
      try {
        const [compResp, gamedaysResp, meResp] = await Promise.all([
          fetch(`/api/clubs/${activeClub.id}/competitions`, {
            credentials: "include",
          }),
          fetch(`/api/clubs/${activeClub.id}/gamedays`, {
            credentials: "include",
          }),
          fetch(`/api/clubs/${activeClub.id}/players/me`, {
            credentials: "include",
          }),
        ]);

        if (cancelled) return;
        if (!compResp.ok || !gamedaysResp.ok) {
          setCompetitions([]);
          setLastGameDayDetail(null);
          setMyPlayer(null);
          setIsLoading(false);
          return;
        }

        const comps: Competition[] = await compResp.json();
        const gamedays: Array<{ id: string; date: string }> =
          await gamedaysResp.json();
        let player: Player | null = null;
        if (meResp.ok) {
          player = await meResp.json();
        }

        setCompetitions(comps);
        setMyPlayer(player);

        if (comps.length === 0 || gamedays.length === 0) {
          setLastGameDayDetail(null);
          setIsLoading(false);
          return;
        }

        const lastId = gamedays[0].id;
        const detailResp = await fetch(
          `/api/clubs/${activeClub.id}/gamedays/${lastId}`,
          { credentials: "include" }
        );
        if (cancelled) return;
        if (detailResp.ok) {
          const detail: GameDayDetail = await detailResp.json();
          setLastGameDayDetail(detail);
        } else {
          setLastGameDayDetail(null);
        }
      } catch (e) {
        if (!cancelled) {
          setError(String(e));
          setCompetitions([]);
          setLastGameDayDetail(null);
          setMyPlayer(null);
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    };

    run();
    return () => {
      cancelled = true;
    };
  }, [activeClub]);

  return {
    competitions,
    lastGameDayDetail,
    myPlayer,
    isLoading,
    error,
  };
}
