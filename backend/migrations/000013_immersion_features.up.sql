-- 1. Fog of War on Maps
ALTER TABLE maps
    ADD COLUMN fog_enabled TINYINT(1) NOT NULL DEFAULT 0 AFTER height,
    ADD COLUMN fog_revealed LONGTEXT NOT NULL AFTER fog_enabled;

-- 2. Area of Effect (AoE) Spell & Hazard Templates
CREATE TABLE aoe_templates (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    map_id     BIGINT UNSIGNED NOT NULL,
    name       VARCHAR(120)    NOT NULL,
    shape      ENUM('circle', 'cone', 'line', 'cube') NOT NULL DEFAULT 'circle',
    size_ft    INT             NOT NULL DEFAULT 20,
    color      VARCHAR(24)     NOT NULL DEFAULT '#e74c3c',
    x          INT             NOT NULL DEFAULT 0,
    y          INT             NOT NULL DEFAULT 0,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_aoe_templates_map (map_id),
    CONSTRAINT fk_aoe_templates_map FOREIGN KEY (map_id) REFERENCES maps (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Campaign Handouts & Secret DM Notes / Lore
CREATE TABLE handouts (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    campaign_id BIGINT UNSIGNED NOT NULL,
    title       VARCHAR(120)    NOT NULL,
    content     TEXT            NOT NULL,
    image_url   VARCHAR(255)    NOT NULL DEFAULT '',
    is_revealed TINYINT(1)      NOT NULL DEFAULT 0,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_handouts_campaign (campaign_id),
    CONSTRAINT fk_handouts_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
