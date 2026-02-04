-- Allow club delete to cascade: when competitions are deleted (via club CASCADE),
-- game_day_competition_values referencing them must be deleted too instead of blocking.
ALTER TABLE game_day_competition_values
  DROP CONSTRAINT game_day_competition_values_competition_id_fkey;

ALTER TABLE game_day_competition_values
  ADD CONSTRAINT game_day_competition_values_competition_id_fkey
  FOREIGN KEY (competition_id) REFERENCES competitions(id) ON DELETE CASCADE;
