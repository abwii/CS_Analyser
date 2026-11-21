-- Remove duplicates that may have accumulated before this constraint existed
-- (e.g. from SQS message redelivery), keeping the lowest id per group.
DELETE FROM player_positions a
USING player_positions b
WHERE a.id > b.id
  AND a.demo_id = b.demo_id
  AND a.steamid = b.steamid
  AND a.round = b.round
  AND a.tick = b.tick;

ALTER TABLE player_positions
    ADD CONSTRAINT player_positions_demo_steamid_round_tick_key
    UNIQUE (demo_id, steamid, round, tick);
