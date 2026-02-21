export type Gender = "male" | "female";

export type Player = {
  id: string;
  club_id: string;
  user_id: string | null;
  role_id: string | null;
  name: string;
  balance: number; // in Cent
  start_balance: number; // in Cent
  gender?: Gender | null;
  inactive: boolean;
  partner_id?: string | null;
  pair_balance?: number | null; // only set when player has partner (own + partner balance)
  created_at: string;
  updated_at: string;
};

export type CreatePlayerRequest = {
  name: string;
  balance: number;
  start_balance: number;
  user_id?: string | null;
  role_id: string; // Required
  gender?: Gender | null;
  inactive?: boolean;
};

export type UpdatePlayerRequest = {
  name: string;
  balance: number;
  start_balance: number;
  user_id?: string | null;
  role_id: string; // Required
  gender?: Gender | null;
  inactive?: boolean;
  partner_id?: string | null; // optional partner (Paar-Modus)
};

// Re-export Role from role.ts to avoid breaking existing imports
export type { Role } from "./role";

// Hilfsfunktion zum Formatieren von Cent in Euro
export function formatCentsToEuro(cents: number): string {
  const euro = cents / 100;
  return new Intl.NumberFormat("de-DE", {
    style: "currency",
    currency: "EUR",
  }).format(euro);
}

// Hilfsfunktion zum Konvertieren von Euro in Cent
export function euroToCents(euro: number): number {
  return Math.round(euro * 100);
}

