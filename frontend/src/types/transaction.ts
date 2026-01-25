export type TransactionType = 'base_fee' | 'fee' | 'deposit' | 'tip' | 'expense';

export interface Transaction {
  id: string;
  club_id: string;
  player_id?: string;
  player_name?: string;
  transaction_type: TransactionType;
  amount: number; // in cents
  description: string;
  game_day_fee_id?: string;
  game_day_id?: string;
  player_balance_before?: number;
  player_balance_after?: number;
  club_balance_before: number;
  club_balance_after: number;
  created_at: string;
  updated_at: string;
}

export interface PaginatedTransactions {
  data: Transaction[];
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface CreateTransactionRequest {
  transaction_type: 'deposit' | 'tip' | 'expense';
  player_id?: string; // Required for deposit
  amount: number; // in cents
  description: string;
}

export interface TransactionSummary {
  base_fee_total: number; // in cents
  base_fee_count: number;
  penalty_fee_total: number; // in cents
  penalty_fee_count: number;
  total: number; // in cents
}

// Helper function to get transaction type label
export function getTransactionTypeLabel(type: TransactionType): string {
  const labels: Record<TransactionType, string> = {
    base_fee: 'Base Fee',
    fee: 'Penalty Fee',
    deposit: 'Deposit',
    tip: 'Tip',
    expense: 'Expense',
  };
  return labels[type];
}

// Helper function to get transaction type color (for badges)
export function getTransactionTypeColor(type: TransactionType): string {
  const colors: Record<TransactionType, string> = {
    base_fee: 'secondary', // gray
    fee: 'warning', // orange
    deposit: 'success', // green
    tip: 'info', // blue
    expense: 'destructive', // red
  };
  return colors[type];
}

// Helper function to determine if transaction type is income (positive for club)
export function isIncomeTransaction(type: TransactionType): boolean {
  return type === 'deposit' || type === 'tip';
}

// Helper function to determine if transaction type is expense (negative for club)
export function isExpenseTransaction(type: TransactionType): boolean {
  return type === 'expense';
}

// Helper function to determine if transaction type is a fee (negative for player)
export function isFeeTransaction(type: TransactionType): boolean {
  return type === 'base_fee' || type === 'fee';
}
