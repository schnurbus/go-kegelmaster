export type PlayerInvitation = {
  id: string;
  player_id: string;
  email: string;
  token: string;
  expires_at: string;
  accepted_at: string | null;
  created_at: string;
};

export type InvitePlayerRequest = {
  email: string;
};

export type InvitationResponse = {
  player_id: string;
  player_name: string;
  club_id: string;
  club_name: string;
  email: string;
  expires_at: string;
};

export type AcceptInvitationResponse = {
  player_id: string;
  message: string;
};
