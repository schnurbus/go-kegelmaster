export interface PenaltyType {
  id: string;
  club_id: string;
  name: string;
  description: string;
  price: number;
  display_order: number;
  allows_decimal_quantity: boolean;
  created_at: string;
  updated_at: string;
  replaced_by_id?: string;
}

