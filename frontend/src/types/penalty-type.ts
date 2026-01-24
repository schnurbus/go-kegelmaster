export interface PenaltyType {
  id: string;
  club_id: string;
  name: string;
  description: string;
  price: number;
  display_order: number;
  created_at: string;
  updated_at: string;
  replaced_by_id?: string;
}

