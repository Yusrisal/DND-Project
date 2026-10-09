-- 1. Character expansion: status, currency, death saves, hit dice, damage modifiers
ALTER TABLE characters
    ADD COLUMN status ENUM('active', 'tavern', 'dead', 'retired') NOT NULL DEFAULT 'active' AFTER user_id,
    ADD COLUMN cp INT NOT NULL DEFAULT 0 AFTER proficiency_bonus,
    ADD COLUMN sp INT NOT NULL DEFAULT 0 AFTER cp,
    ADD COLUMN ep INT NOT NULL DEFAULT 0 AFTER sp,
    ADD COLUMN gp INT NOT NULL DEFAULT 0 AFTER ep,
    ADD COLUMN pp INT NOT NULL DEFAULT 0 AFTER gp,
    ADD COLUMN death_save_successes INT NOT NULL DEFAULT 0 AFTER notes,
    ADD COLUMN death_save_failures INT NOT NULL DEFAULT 0 AFTER death_save_successes,
    ADD COLUMN hit_dice_total INT NOT NULL DEFAULT 1 AFTER death_save_failures,
    ADD COLUMN hit_dice_used INT NOT NULL DEFAULT 0 AFTER hit_dice_total,
    ADD COLUMN hit_dice_die VARCHAR(10) NOT NULL DEFAULT 'd8' AFTER hit_dice_used,
    ADD COLUMN damage_resistances VARCHAR(255) NOT NULL DEFAULT '' AFTER hit_dice_die,
    ADD COLUMN damage_vulnerabilities VARCHAR(255) NOT NULL DEFAULT '' AFTER damage_resistances,
    ADD COLUMN damage_immunities VARCHAR(255) NOT NULL DEFAULT '' AFTER damage_vulnerabilities,
    ADD COLUMN condition_immunities VARCHAR(255) NOT NULL DEFAULT '' AFTER damage_immunities;

-- 2. Items cost in copper pieces
ALTER TABLE character_items
    ADD COLUMN cost_cp INT NOT NULL DEFAULT 0 AFTER weight;

-- 3. Monster damage and condition modifiers
ALTER TABLE monsters
    ADD COLUMN damage_resistances VARCHAR(255) NOT NULL DEFAULT '' AFTER traits,
    ADD COLUMN damage_vulnerabilities VARCHAR(255) NOT NULL DEFAULT '' AFTER damage_resistances,
    ADD COLUMN damage_immunities VARCHAR(255) NOT NULL DEFAULT '' AFTER damage_vulnerabilities,
    ADD COLUMN condition_immunities VARCHAR(255) NOT NULL DEFAULT '' AFTER damage_immunities;

-- 4. Shops and Shop Items
CREATE TABLE shops (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    campaign_id BIGINT UNSIGNED NOT NULL,
    name        VARCHAR(120)    NOT NULL,
    description TEXT            NOT NULL,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_shops_campaign (campaign_id),
    CONSTRAINT fk_shops_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE shop_items (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    shop_id      BIGINT UNSIGNED NOT NULL,
    name         VARCHAR(120)    NOT NULL,
    cost_cp      INT             NOT NULL DEFAULT 0,
    quantity     INT             NOT NULL DEFAULT -1,
    weight       DECIMAL(8,2)    NOT NULL DEFAULT 0,
    attack_bonus INT             NOT NULL DEFAULT 0,
    damage       VARCHAR(40)     NOT NULL DEFAULT '',
    damage_type  VARCHAR(24)     NOT NULL DEFAULT '',
    description  TEXT            NOT NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_shop_items_shop (shop_id),
    CONSTRAINT fk_shop_items_shop FOREIGN KEY (shop_id) REFERENCES shops (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
