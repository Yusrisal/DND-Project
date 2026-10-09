CREATE TABLE characters (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    campaign_id       BIGINT UNSIGNED NOT NULL,
    user_id           BIGINT UNSIGNED NULL,
    name              VARCHAR(120)    NOT NULL,
    class             VARCHAR(60)     NOT NULL DEFAULT '',
    race              VARCHAR(60)     NOT NULL DEFAULT '',
    level             INT             NOT NULL DEFAULT 1,
    hp_current        INT             NOT NULL DEFAULT 0,
    hp_max            INT             NOT NULL DEFAULT 0,
    hp_temp           INT             NOT NULL DEFAULT 0,
    armor_class       INT             NOT NULL DEFAULT 10,
    speed             INT             NOT NULL DEFAULT 30,
    proficiency_bonus INT             NOT NULL DEFAULT 2,
    strength          INT             NOT NULL DEFAULT 10,
    dexterity         INT             NOT NULL DEFAULT 10,
    constitution      INT             NOT NULL DEFAULT 10,
    intelligence      INT             NOT NULL DEFAULT 10,
    wisdom            INT             NOT NULL DEFAULT 10,
    charisma          INT             NOT NULL DEFAULT 10,
    notes             TEXT            NOT NULL,
    created_at        DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_characters_campaign (campaign_id),
    KEY idx_characters_user (user_id),
    CONSTRAINT fk_characters_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE CASCADE,
    CONSTRAINT fk_characters_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE character_items (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    character_id BIGINT UNSIGNED NOT NULL,
    name         VARCHAR(120)    NOT NULL,
    quantity     INT             NOT NULL DEFAULT 1,
    weight       DECIMAL(8,2)    NOT NULL DEFAULT 0,
    equipped     TINYINT(1)      NOT NULL DEFAULT 0,
    description  TEXT            NOT NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_character_items_character (character_id),
    CONSTRAINT fk_character_items_character FOREIGN KEY (character_id) REFERENCES characters (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE character_spells (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    character_id BIGINT UNSIGNED NOT NULL,
    name         VARCHAR(120)    NOT NULL,
    level        INT             NOT NULL DEFAULT 0,
    school       VARCHAR(40)     NOT NULL DEFAULT '',
    prepared     TINYINT(1)      NOT NULL DEFAULT 0,
    description  TEXT            NOT NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_character_spells_character (character_id, level),
    CONSTRAINT fk_character_spells_character FOREIGN KEY (character_id) REFERENCES characters (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE character_spell_slots (
    character_id BIGINT UNSIGNED NOT NULL,
    level        INT             NOT NULL,
    total        INT             NOT NULL DEFAULT 0,
    used         INT             NOT NULL DEFAULT 0,
    PRIMARY KEY (character_id, level),
    CONSTRAINT fk_character_spell_slots_character FOREIGN KEY (character_id) REFERENCES characters (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE dice_rolls (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    campaign_id  BIGINT UNSIGNED NOT NULL,
    user_id      BIGINT UNSIGNED NULL,
    character_id BIGINT UNSIGNED NULL,
    author_name  VARCHAR(80)     NOT NULL,
    label        VARCHAR(120)    NOT NULL DEFAULT '',
    expression   VARCHAR(120)    NOT NULL,
    detail       JSON            NOT NULL,
    total        INT             NOT NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_dice_rolls_campaign (campaign_id, id),
    CONSTRAINT fk_dice_rolls_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE CASCADE,
    CONSTRAINT fk_dice_rolls_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT fk_dice_rolls_character FOREIGN KEY (character_id) REFERENCES characters (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
