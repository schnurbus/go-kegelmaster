import { useEffect, useRef } from "react";

import { useSidebar } from "@/components/ui/sidebar";
import { useAuth } from "@/context/AuthContext";
import { useClub } from "@/context/ClubContext";
import { useTour } from "@/context/TourContext";
import { usePermissions } from "@/hooks/use-permissions";
import { usePublicConfig } from "@/hooks/use-public-config";
import { takeTourTrigger } from "@/lib/onboarding";

import { runTour, type TourHandle } from "./run-tour";
import { stepsForFlow, type TourContextInfo, type TourFlow } from "./tour-catalog";

type Plan = {
  flow: TourFlow;
  /** Nach dem Ende „Rundgang gesehen“ setzen. */
  markTour: boolean;
  /** Nach dem Ende „Hinweis zur Ersteinrichtung gesehen“ setzen. */
  markSetupHint: boolean;
};

type TourAutoStartProps = {
  /** Stand der Ersteinrichtung ist geladen (bzw. der Benutzer ist nicht Besitzer). */
  setupReady: boolean;
  showSetupCard: boolean;
};

/**
 * Entscheidet auf dem Dashboard, ob ein Rundgang oder ein einzelner Hinweis
 * startet, und merkt danach am Konto, was gesehen wurde.
 */
export function TourAutoStart({ setupReady, showSetupCard }: TourAutoStartProps) {
  const { user, updateSettings } = useAuth();
  const { activeClub } = useClub();
  const { isMobile, setOpenMobile } = useSidebar();
  const { manualRequest } = useTour();
  const { isLoading: permsLoading, canList, canView } = usePermissions(activeClub?.id ?? null);
  const config = usePublicConfig();

  const isOwner = !!user && !!activeClub && activeClub.user_id === user.id;
  const ctx: TourContextInfo = {
    clubName: activeClub?.name ?? "",
    isOwner,
    canSeeRoles: isOwner || canList("roles") || canView("roles"),
    chatEnabled: !!config?.chat_enabled,
    showSetupCard,
  };
  const ctxRef = useRef(ctx);
  ctxRef.current = ctx;

  const ready = !!user && !!activeClub && !permsLoading && setupReady;
  const planRef = useRef<Plan | null | undefined>(undefined);
  const handledManualRef = useRef(manualRequest);
  const handleRef = useRef<TourHandle | null>(null);

  useEffect(() => {
    if (!ready || !user) return;
    const ctxNow = ctxRef.current;

    if (manualRequest !== handledManualRef.current) {
      // Manually started: ignores the flags.
      handledManualRef.current = manualRequest;
      planRef.current = { flow: "full", markTour: true, markSetupHint: ctxNow.showSetupCard };
    } else if (planRef.current === undefined) {
      planRef.current = decide(user.tour_seen_at, user.setup_hint_seen_at, ctxNow);
    }
    const plan = planRef.current;
    if (!plan) return;

    const steps = stepsForFlow(plan.flow, ctxNow);
    const includesSetup = steps.some((s) => s.id === "setup-card");
    handleRef.current?.cancel();
    handleRef.current = runTour({
      steps,
      ctx: ctxNow,
      isMobile,
      setSidebarOpen: setOpenMobile,
      onEnd: () => {
        planRef.current = null;
        const payload = {
          ...(plan.markTour && { tour_seen: true }),
          ...((plan.markSetupHint || includesSetup) && { setup_hint_seen: true }),
        };
        if (Object.keys(payload).length > 0) {
          updateSettings(payload).catch((err) => console.error("Rundgang-Merker setzen:", err));
        }
      },
    });

    return () => {
      // Leaving the dashboard (or React StrictMode re-running the effect) stops
      // the tour without marking it as seen; the plan is kept for a restart.
      handleRef.current?.cancel();
      handleRef.current = null;
    };
    // ctx is read through ctxRef: later changes must not restart a running tour.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, manualRequest]);

  return null;
}

function decide(tourSeenAt: string | null, setupHintSeenAt: string | null, ctx: TourContextInfo): Plan | null {
  const trigger = takeTourTrigger();
  const fullTour: Plan = { flow: "full", markTour: true, markSetupHint: false };

  if (trigger === "invitation-accepted") {
    return tourSeenAt ? { flow: "club-switcher-hint", markTour: false, markSetupHint: false } : fullTour;
  }
  if (!tourSeenAt) {
    return fullTour;
  }
  if (trigger === "club-created" && !setupHintSeenAt && ctx.isOwner && ctx.showSetupCard) {
    return { flow: "setup-hint", markTour: false, markSetupHint: true };
  }
  return null;
}
