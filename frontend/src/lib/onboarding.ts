// Anlass, aus dem das Dashboard beim nächsten Öffnen einen Rundgang oder einen
// einzelnen Hinweis zeigt. Liegt im sessionStorage, damit er einen Seitenwechsel
// übersteht, aber nicht in andere Tabs oder spätere Sitzungen wandert.
export type TourTrigger = "club-created" | "invitation-accepted";

const PENDING_KEY = "kegelmaster-tour-pending";

export function setTourTrigger(trigger: TourTrigger) {
  try {
    sessionStorage.setItem(PENDING_KEY, trigger);
  } catch {
    // Ohne Storage gibt es eben keinen Hinweis.
  }
}

/** Liefert den vorgemerkten Anlass und löscht ihn (einmalig). */
export function takeTourTrigger(): TourTrigger | null {
  try {
    const value = sessionStorage.getItem(PENDING_KEY);
    sessionStorage.removeItem(PENDING_KEY);
    return value === "club-created" || value === "invitation-accepted" ? value : null;
  } catch {
    return null;
  }
}
