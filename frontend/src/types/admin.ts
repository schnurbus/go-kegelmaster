// Types of the instance admin API (/api/admin), see openapi.yaml (tag "admin").

export type AdminStats = {
  users: {
    total: number;
    new_7d: number;
    new_30d: number;
    active_30d: number;
    instance_admins: number;
    disabled: number;
  };
  clubs: { total: number; active_90d: number };
  players: { total: number; linked: number };
  open_invitations: number;
  api_tokens: { active: number };
  passkeys: number;
  anomalies: {
    orphaned_users: number;
    inactive_clubs: number;
    api_tokens_without_expiry: number;
    api_tokens_unused_90d: number;
    registration_spikes: { date: string; registrations: number }[];
  };
};

export type AdminPage<T> = {
  data: T[];
  page: number;
  limit: number;
  total: number;
  total_pages: number;
};

export type AdminUserRef = { id: string; email: string };

export type AdminUserListItem = {
  id: string;
  email: string;
  is_instance_admin: boolean;
  orphaned: boolean;
  created_at: string;
  last_login_at: string | null;
  /** Set while the account is disabled. */
  disabled_at: string | null;
  owned_club_count: number;
  player_club_count: number;
  passkey_count: number;
  api_token_count: number;
};

export type AdminUserDetail = {
  id: string;
  email: string;
  is_instance_admin: boolean;
  orphaned: boolean;
  created_at: string;
  updated_at: string;
  password_changed_at: string | null;
  last_login_at: string | null;
  /** Set while the account is disabled. */
  disabled_at: string | null;
  /** Browser sessions issued before this were revoked by an admin. */
  sessions_valid_after: string | null;
  owned_clubs: { id: string; name: string; created_at: string; player_count: number }[];
  players: {
    player_id: string;
    player_name: string;
    inactive: boolean;
    club_id: string;
    club_name: string;
    is_owner: boolean;
    role_id: string | null;
    role_name: string | null;
  }[];
  api_tokens: {
    id: string;
    name: string;
    token_prefix: string;
    created_at: string;
    expires_at: string | null;
    last_used_at: string | null;
  }[];
  passkeys: { id: string; name: string; created_at: string; last_used_at: string | null }[];
};

export type AdminClubListItem = {
  id: string;
  name: string;
  created_at: string;
  owner: AdminUserRef;
  player_count: number;
  linked_player_count: number;
  game_day_count: number;
  last_activity_at: string | null;
};

export type AdminClubDetail = AdminClubListItem & {
  transaction_count: number;
  players: {
    id: string;
    name: string;
    inactive: boolean;
    created_at: string;
    user: AdminUserRef | null;
    role_id: string | null;
    role_name: string | null;
  }[];
  open_invitations: {
    id: string;
    player_id: string;
    player_name: string;
    email: string;
    created_at: string;
    expires_at: string;
  }[];
  roles: { id: string; name: string; player_count: number }[];
};
