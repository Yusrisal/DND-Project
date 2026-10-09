DROP TABLE IF EXISTS campaign_savepoints;
ALTER TABLE campaigns
    DROP COLUMN pause_message,
    DROP COLUMN is_paused;
