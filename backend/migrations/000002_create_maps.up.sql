CREATE TABLE maps (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    campaign_id BIGINT UNSIGNED NOT NULL,
    name        VARCHAR(120)    NOT NULL,
    image_url   VARCHAR(512)    NOT NULL DEFAULT '',
    grid_size   INT             NOT NULL DEFAULT 50,
    width       INT             NOT NULL DEFAULT 1920,
    height      INT             NOT NULL DEFAULT 1080,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_maps_campaign_id (campaign_id),
    CONSTRAINT fk_maps_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
