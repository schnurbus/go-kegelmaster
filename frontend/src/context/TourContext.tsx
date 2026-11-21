import type { ReactNode } from "react";
import { createContext, useCallback, useContext, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";

type TourContextValue = {
  /** Zählt manuelle Starts hoch; das Dashboard startet bei jeder Änderung den vollen Rundgang. */
  manualRequest: number;
  /** „Rundgang erneut starten“: öffnet das Dashboard und startet den Rundgang, unabhängig vom Merker. */
  restartTour: () => void;
};

const TourContext = createContext<TourContextValue | undefined>(undefined);

export function TourProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate();
  const [manualRequest, setManualRequest] = useState(0);

  const restartTour = useCallback(() => {
    navigate("/app");
    setManualRequest((n) => n + 1);
  }, [navigate]);

  const value = useMemo(() => ({ manualRequest, restartTour }), [manualRequest, restartTour]);
  return <TourContext.Provider value={value}>{children}</TourContext.Provider>;
}

export function useTour() {
  const ctx = useContext(TourContext);
  if (!ctx) {
    throw new Error("useTour must be used within TourProvider");
  }
  return ctx;
}
