-- 000021_core_conditions.down.sql
--
-- Only the seeded rows. Anything a DM added gets a higher id and survives.
DELETE FROM status_effects WHERE id >= 1 AND id <= 15 AND campaign_id IS NULL;
