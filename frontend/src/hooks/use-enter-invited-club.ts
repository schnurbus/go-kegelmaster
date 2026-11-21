import { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { useClub } from "@/context/ClubContext";
import type { AcceptedInvitation } from "@/lib/invitation";
import { setTourTrigger } from "@/lib/onboarding";

/**
 * Nach einer angenommenen Einladung: den eingeladenen Club aktiv setzen und auf
 * dessen Dashboard wechseln. Dort startet der Rundgang bzw. der Hinweis auf
 * den Club-Switcher.
 */
export function useEnterInvitedClub() {
  const { refreshClubs } = useClub();
  const navigate = useNavigate();

  return useCallback(
    async (accepted: AcceptedInvitation) => {
      setTourTrigger("invitation-accepted");
      await refreshClubs(accepted.club_id);
      navigate("/app", { replace: true });
    },
    [refreshClubs, navigate],
  );
}
