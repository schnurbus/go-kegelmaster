import { useCallback, useEffect, useState } from "react";

export type SetupStepId =
  | "club"
  | "penalty_types"
  | "roles"
  | "players"
  | "first_game_day"
  | "own_player"
  | "invitations"
  | "competitions";

export type SetupStep = { id: SetupStepId; required: boolean; done: boolean };

/** Stand der Ersteinrichtung aus `GET /api/clubs/{id}/setup` (nur für den Besitzer). */
export type ClubSetup = { dismissed: boolean; complete: boolean; steps: SetupStep[] };

/**
 * Lädt den Stand der Ersteinrichtung, wenn enabled (der Benutzer ist Besitzer).
 * ready ist true, sobald feststeht, ob die Karte zu sehen ist.
 */
export function useClubSetup(clubId: string | null, enabled: boolean) {
  const [setup, setSetup] = useState<ClubSetup | null>(null);
  const [loadedFor, setLoadedFor] = useState<string | null>(null);

  const reload = useCallback(async () => {
    if (!clubId || !enabled) return;
    try {
      const resp = await fetch(`/api/clubs/${clubId}/setup`, { credentials: "include" });
      setSetup(resp.ok ? ((await resp.json()) as ClubSetup) : null);
    } catch {
      setSetup(null);
    } finally {
      setLoadedFor(clubId);
    }
  }, [clubId, enabled]);

  useEffect(() => {
    setSetup(null);
    setLoadedFor(null);
    void reload();
  }, [reload]);

  const ready = !enabled || (!!clubId && loadedFor === clubId);
  const visible = enabled && !!setup && !setup.dismissed && !setup.complete;
  return { setup: enabled ? setup : null, ready, visible, reload };
}
