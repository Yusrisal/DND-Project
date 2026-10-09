-- 1. Weather and Atmosphere on Maps
ALTER TABLE maps
    ADD COLUMN weather VARCHAR(32) NOT NULL DEFAULT 'clear' AFTER fog_revealed;

-- 2. Interactive Map Props (Doors, Chests, Traps, Clue Markers)
CREATE TABLE map_objects (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    map_id     BIGINT UNSIGNED NOT NULL,
    kind       ENUM('door', 'chest', 'trap', 'marker') NOT NULL DEFAULT 'door',
    name       VARCHAR(120)    NOT NULL,
    x          INT             NOT NULL DEFAULT 0,
    y          INT             NOT NULL DEFAULT 0,
    state      VARCHAR(40)     NOT NULL DEFAULT 'closed',
    data       JSON            NOT NULL,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_map_objects_map (map_id),
    CONSTRAINT fk_map_objects_map FOREIGN KEY (map_id) REFERENCES maps (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
