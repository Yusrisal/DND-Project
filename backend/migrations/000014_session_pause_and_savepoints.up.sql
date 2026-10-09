-- 1. Campaign live session pause
ALTER TABLE campaigns
    ADD COLUMN is_paused TINYINT(1) NOT NULL DEFAULT 0 AFTER description,
    ADD COLUMN pause_message VARCHAR(255) NOT NULL DEFAULT 'Session Paused by DM' AFTER is_paused;

-- 2. Campaign Savepoints / State Snapshots
CREATE TABLE campaign_savepoints (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    campaign_id   BIGINT UNSIGNED NOT NULL,
    map_id        BIGINT UNSIGNED NULL,
    name          VARCHAR(120)    NOT NULL,
    notes         TEXT            NOT NULL,
    snapshot_data LONGTEXT        NOT NULL,
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_campaign_savepoints_campaign (campaign_id),
    KEY idx_campaign_savepoints_map (map_id),
    CONSTRAINT fk_campaign_savepoints_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE CASCADE,
    CONSTRAINT fk_campaign_savepoints_map FOREIGN KEY (map_id) REFERENCES maps (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
