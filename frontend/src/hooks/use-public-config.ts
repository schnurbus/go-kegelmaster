import { useEffect, useState } from "react";

/** Wer ein Konto anlegen darf (REGISTRATION_MODE im Backend). */
export type RegistrationMode = "open" | "invite_only" | "closed";

export type PublicConfig = {
  version?: string;
  chat_enabled?: boolean;
  buy_me_a_coffee_url?: string;
  registration_mode?: RegistrationMode;
};

// One request per page load; all components share the result.
let cached: Promise<PublicConfig | null> | null = null;

function loadPublicConfig(): Promise<PublicConfig | null> {
  if (!cached) {
    cached = fetch("/api/config/public")
      .then((res) => (res.ok ? (res.json() as Promise<PublicConfig>) : null))
      .catch(() => null)
      .then((data) => {
        // Retry on the next use if the request failed.
        if (data === null) cached = null;
        return data;
      });
  }
  return cached;
}

/** Öffentliche Konfiguration aus `GET /api/config/public`; `null` solange sie lädt oder nicht verfügbar ist. */
export function usePublicConfig(): PublicConfig | null {
  const [config, setConfig] = useState<PublicConfig | null>(null);

  useEffect(() => {
    let cancelled = false;
    loadPublicConfig().then((data) => {
      if (!cancelled) setConfig(data);
    });
    return () => {
      cancelled = true;
    };
  }, []);

  return config;
}

/**
 * Registrierungsmodus der Instanz. Solange die Konfiguration lädt (oder ein
 * älteres Backend nichts liefert), gilt wie bisher `open`.
 */
export function useRegistrationMode(): RegistrationMode {
  return usePublicConfig()?.registration_mode ?? "open";
}
