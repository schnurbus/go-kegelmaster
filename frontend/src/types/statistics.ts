// Response of GET /api/clubs/:clubId/players/me/statistics (amounts in cents).

export interface StatisticsGameDayAmount {
  game_day_id: string;
  date: string;
  amount: number;
}

export interface PenaltyTypeSummary {
  penalty_type_id: string;
  name: string;
  quantity: number;
  amount: number;
  game_days: number;
  average_per_game_day: number;
  club_average_per_game_day: number;
}

export interface CompetitionHistoryEntry {
  game_day_id: string;
  date: string;
  value: number;
  club_average: number;
  is_win: boolean;
  is_loss: boolean;
}

export interface CompetitionStatistics {
  competition_id: string;
  name: string;
  scoring_type: "winner" | "loser" | "both";
  is_gender_specific: boolean;
  participations: number;
  total: number;
  average: number;
  club_average: number;
  best: number | null;
  worst: number | null;
  wins: number;
  losses: number;
  rank: number;
  rank_of: number;
  history: CompetitionHistoryEntry[];
}

export interface PlayerStatistics {
  player: { id: string; name: string };
  period: { from: string | null; to: string };
  attendance: {
    club_game_days: number;
    attended: number;
    rate: number;
    current_streak: number;
    longest_streak: number;
    first_attended: string | null;
    last_attended: string | null;
  };
  penalties: {
    total_amount: number;
    absent_amount: number; // part of total_amount booked while absent
    average_per_game_day: number;
    club_average_per_game_day: number;
    penalty_free_game_days: number;
    most_expensive_game_day: StatisticsGameDayAmount | null;
    cheapest_game_day: StatisticsGameDayAmount | null;
    rank: number;
    rank_of: number;
    by_type: PenaltyTypeSummary[];
  };
  competitions: CompetitionStatistics[];
  finance: {
    start_balance: number;
    end_balance: number;
    deposits: number;
    fees: number;
    base_fees: number;
    balance_history: Array<{ date: string; balance: number }>;
  };
  game_days: Array<{
    game_day_id: string;
    date: string;
    attended: boolean;
    participants: number;
    penalty_amount: number | null;
    club_average_amount: number;
  }>;
}
