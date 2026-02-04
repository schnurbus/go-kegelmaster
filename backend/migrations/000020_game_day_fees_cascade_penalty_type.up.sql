-- Allow club delete to cascade: when penalty_types are deleted (via club CASCADE),
-- game_day_fees referencing them must be deleted too instead of blocking.
ALTER TABLE game_day_fees
  DROP CONSTRAINT game_day_fees_penalty_type_id_fkey;

ALTER TABLE game_day_fees
  ADD CONSTRAINT game_day_fees_penalty_type_id_fkey
  FOREIGN KEY (penalty_type_id) REFERENCES penalty_types(id) ON DELETE CASCADE;
