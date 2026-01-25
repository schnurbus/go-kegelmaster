export interface GameDay {
  id: string;
  club_id: string;
  date: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface GameDaySummary extends GameDay {
  participant_count: number;
  penalty_fee_total: number; // in cents
}

export function formatCentsToEuro(cents: number): string {
  return `${(Math.abs(cents) / 100).toFixed(2)} €`;
}
