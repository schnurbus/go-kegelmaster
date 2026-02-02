export type ScoringType = "winner" | "loser" | "both";

export interface Competition {
  id: string;
  club_id: string;
  name: string;
  scoring_type: ScoringType;
  is_gender_specific: boolean;
  display_order: number;
  created_at: string;
  updated_at: string;
}

export type CreateCompetitionRequest = {
  name: string;
  scoring_type: ScoringType;
  is_gender_specific: boolean;
  display_order?: number;
};

export type UpdateCompetitionRequest = {
  name: string;
  scoring_type: ScoringType;
  is_gender_specific: boolean;
  display_order: number;
};

export interface GameDayCompetitionValue {
  id: string;
  game_day_participant_id: string;
  competition_id: string;
  value: number;
  created_at: string;
  updated_at: string;
}
