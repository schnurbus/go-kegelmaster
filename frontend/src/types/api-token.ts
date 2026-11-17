export type ApiToken = {
  id: string;
  name: string;
  token_prefix: string;
  expires_at: string | null;
  last_used_at: string | null;
  created_at: string;
};

export type CreatedApiToken = ApiToken & {
  token: string;
};
