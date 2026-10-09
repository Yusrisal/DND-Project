-- Environments: weather, terrain, traps, hazard zones
CREATE TABLE environments (
    id                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    campaign_id        BIGINT UNSIGNED NOT NULL,
    map_id             BIGINT UNSIGNED NULL,
    name               VARCHAR(120)    NOT NULL,
    kind               VARCHAR(40)     NOT NULL DEFAULT 'hazard',
    difficulty_dc      INT             NOT NULL DEFAULT 0,
    damage             VARCHAR(40)     NOT NULL DEFAULT '',
    damage_type        VARCHAR(24)     NOT NULL DEFAULT '',
    effect_description TEXT            NOT NULL,
    active             TINYINT(1)      NOT NULL DEFAULT 1,
    created_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_environments_campaign (campaign_id),
    KEY idx_environments_map (map_id),
    CONSTRAINT fk_environments_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE CASCADE,
    CONSTRAINT fk_environments_map FOREIGN KEY (map_id) REFERENCES maps (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Compendium items: global/campaign items repository
CREATE TABLE compendium_items (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    campaign_id  BIGINT UNSIGNED NULL,
    name         VARCHAR(120)    NOT NULL,
    category     VARCHAR(40)     NOT NULL DEFAULT 'gear',
    rarity       VARCHAR(30)     NOT NULL DEFAULT 'common',
    cost_cp      INT             NOT NULL DEFAULT 0,
    weight       DECIMAL(8,2)    NOT NULL DEFAULT 0,
    attack_bonus INT             NOT NULL DEFAULT 0,
    damage       VARCHAR(40)     NOT NULL DEFAULT '',
    damage_type  VARCHAR(24)     NOT NULL DEFAULT '',
    description  TEXT            NOT NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_compendium_items_campaign (campaign_id),
    KEY idx_compendium_items_name (name),
    CONSTRAINT fk_compendium_items_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Monster drop tables: loot dropped on death/harvest
CREATE TABLE monster_drops (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    monster_id   BIGINT UNSIGNED NOT NULL,
    item_name    VARCHAR(120)    NOT NULL,
    item_id      BIGINT UNSIGNED NULL,
    drop_rate    INT             NOT NULL DEFAULT 100,
    min_quantity INT             NOT NULL DEFAULT 1,
    max_quantity INT             NOT NULL DEFAULT 1,
    cp           INT             NOT NULL DEFAULT 0,
    sp           INT             NOT NULL DEFAULT 0,
    gp           INT             NOT NULL DEFAULT 0,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_monster_drops_monster (monster_id),
    CONSTRAINT fk_monster_drops_monster FOREIGN KEY (monster_id) REFERENCES monsters (id) ON DELETE CASCADE,
    CONSTRAINT fk_monster_drops_item FOREIGN KEY (item_id) REFERENCES compendium_items (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Compendium spells & skills: reference library
CREATE TABLE compendium_spells (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    campaign_id  BIGINT UNSIGNED NULL,
    name         VARCHAR(120)    NOT NULL,
    level        INT             NOT NULL DEFAULT 0,
    school       VARCHAR(40)     NOT NULL DEFAULT '',
    casting_time VARCHAR(40)     NOT NULL DEFAULT '1 action',
    range_area   VARCHAR(60)     NOT NULL DEFAULT '',
    duration     VARCHAR(60)     NOT NULL DEFAULT 'instantaneous',
    damage       VARCHAR(40)     NOT NULL DEFAULT '',
    damage_type  VARCHAR(24)     NOT NULL DEFAULT '',
    save_ability VARCHAR(20)     NOT NULL DEFAULT '',
    description  TEXT            NOT NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_compendium_spells_campaign (campaign_id),
    KEY idx_compendium_spells_name (name),
    CONSTRAINT fk_compendium_spells_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
