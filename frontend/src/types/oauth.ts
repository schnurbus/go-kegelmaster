/** Pending authorization request shown on the consent page. */
export type OAuthConsentInfo = {
  client_id: string;
  client_name: string;
  redirect_uri: string;
  redirect_host: string;
  scope: string;
};

/** A connected application (OAuth grant) of the current user. */
export type OAuthGrant = {
  id: string;
  client_id: string;
  client_name: string;
  redirect_host: string;
  scope: string;
  created_at: string;
  last_used_at: string | null;
};
