import { useEffect } from "react";
import { useSearchParams } from "react-router-dom";

/**
 * Öffnet einen Dialog, wenn die Seite mit ?<param>=… aufgerufen wurde (z. B. aus
 * der Ersteinrichtung), und entfernt den Parameter danach wieder aus der URL.
 * enabled verzögert das Öffnen, bis die Seite bereit ist (z. B. Rollen geladen).
 */
export function useOpenFromQuery(param: string, onOpen: (value: string) => void, enabled = true) {
  const [searchParams, setSearchParams] = useSearchParams();
  const value = searchParams.get(param);

  useEffect(() => {
    if (!enabled || value === null) return;
    onOpen(value);
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.delete(param);
        return next;
      },
      { replace: true },
    );
    // onOpen is an inline callback; the dialog must open only once per value.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled, value, param, setSearchParams]);
}
