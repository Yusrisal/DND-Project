-- Tokens gain their own vitals so a monster can be attacked without a sheet.
-- When character_id is set the sheet wins, and these columns are ignored.
ALTER TABLE tokens
    ADD COLUMN character_id BIGINT UNSIGNED NULL AFTER owner_user_id,
    ADD COLUMN hp_current INT NOT NULL DEFAULT 0 AFTER size,
    ADD COLUMN hp_max INT NOT NULL DEFAULT 0 AFTER hp_current,
    ADD COLUMN armor_class INT NOT NULL DEFAULT 10 AFTER hp_max,
    ADD KEY idx_tokens_character (character_id),
    ADD CONSTRAINT fk_tokens_character FOREIGN KEY (character_id) REFERENCES characters (id) ON DELETE SET NULL;

-- An item with a damage expression is a weapon. Keeping them in one table means
-- the inventory stays one list rather than two that can disagree.
ALTER TABLE character_items
    ADD COLUMN attack_bonus INT NOT NULL DEFAULT 0 AFTER equipped,
    ADD COLUMN damage VARCHAR(40) NOT NULL DEFAULT '' AFTER attack_bonus,
    ADD COLUMN damage_type VARCHAR(24) NOT NULL DEFAULT '' AFTER damage;

CREATE TABLE token_conditions (
    token_id  BIGINT UNSIGNED NOT NULL,
    condition_name VARCHAR(24) NOT NULL,
    PRIMARY KEY (token_id, condition_name),
    CONSTRAINT fk_token_conditions_token FOREIGN KEY (token_id) REFERENCES tokens (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE character_proficiencies (
    character_id BIGINT UNSIGNED NOT NULL,
    kind         ENUM('skill','save') NOT NULL,
    name         VARCHAR(24)     NOT NULL,
    expertise    TINYINT(1)      NOT NULL DEFAULT 0,
    PRIMARY KEY (character_id, kind, name),
    CONSTRAINT fk_character_proficiencies_character FOREIGN KEY (character_id) REFERENCES characters (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- One encounter per map at a time - active is cleared when combat ends.
CREATE TABLE encounters (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    map_id     BIGINT UNSIGNED NOT NULL,
    round      INT             NOT NULL DEFAULT 1,
    turn_index INT             NOT NULL DEFAULT 0,
    active     TINYINT(1)      NOT NULL DEFAULT 1,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_encounters_map (map_id, active),
    CONSTRAINT fk_encounters_map FOREIGN KEY (map_id) REFERENCES maps (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE combatants (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    encounter_id BIGINT UNSIGNED NOT NULL,
    token_id     BIGINT UNSIGNED NULL,
    name         VARCHAR(120)    NOT NULL,
    initiative   INT             NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    KEY idx_combatants_encounter (encounter_id, initiative),
    CONSTRAINT fk_combatants_encounter FOREIGN KEY (encounter_id) REFERENCES encounters (id) ON DELETE CASCADE,
    CONSTRAINT fk_combatants_token FOREIGN KEY (token_id) REFERENCES tokens (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
