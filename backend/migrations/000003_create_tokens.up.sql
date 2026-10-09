CREATE TABLE tokens (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    map_id     BIGINT UNSIGNED NOT NULL,
    name       VARCHAR(120)    NOT NULL,
    image_url  VARCHAR(512)    NOT NULL DEFAULT '',
    color      VARCHAR(16)     NOT NULL DEFAULT '#c0392b',
    size       INT             NOT NULL DEFAULT 1,
    x          DOUBLE          NOT NULL DEFAULT 0,
    y          DOUBLE          NOT NULL DEFAULT 0,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_tokens_map_id (map_id),
    CONSTRAINT fk_tokens_map FOREIGN KEY (map_id) REFERENCES maps (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
