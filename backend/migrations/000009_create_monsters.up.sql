-- The bestiary. A monster here is a template - spawning copies its vitals onto a
-- token so each goblin takes its own damage, while monster_id keeps the link
-- back to the stat block for actions and skills.
CREATE TABLE monsters (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name             VARCHAR(120)    NOT NULL,
    size             VARCHAR(20)     NOT NULL DEFAULT 'medium',
    type             VARCHAR(40)     NOT NULL DEFAULT '',
    alignment        VARCHAR(40)     NOT NULL DEFAULT '',
    armor_class      INT             NOT NULL DEFAULT 10,
    hp_average       INT             NOT NULL DEFAULT 1,
    hp_formula       VARCHAR(40)     NOT NULL DEFAULT '',
    speed            INT             NOT NULL DEFAULT 30,
    strength         INT             NOT NULL DEFAULT 10,
    dexterity        INT             NOT NULL DEFAULT 10,
    constitution     INT             NOT NULL DEFAULT 10,
    intelligence     INT             NOT NULL DEFAULT 10,
    wisdom           INT             NOT NULL DEFAULT 10,
    charisma         INT             NOT NULL DEFAULT 10,
    challenge_rating VARCHAR(10)     NOT NULL DEFAULT '0',
    senses           VARCHAR(191)    NOT NULL DEFAULT '',
    languages        VARCHAR(191)    NOT NULL DEFAULT '',
    traits           TEXT            NOT NULL,
    color            VARCHAR(16)     NOT NULL DEFAULT '#8e5a3b',
    image_url        VARCHAR(512)    NOT NULL DEFAULT '',
    token_size       INT             NOT NULL DEFAULT 1,
    created_at       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_monsters_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Stat blocks print skills and saves as flat modifiers ("Perception +4"), not as
-- a proficiency flag, so that is what is stored.
CREATE TABLE monster_proficiencies (
    monster_id BIGINT UNSIGNED NOT NULL,
    kind       ENUM('skill','save') NOT NULL,
    name       VARCHAR(24)     NOT NULL,
    modifier   INT             NOT NULL DEFAULT 0,
    PRIMARY KEY (monster_id, kind, name),
    CONSTRAINT fk_monster_proficiencies_monster FOREIGN KEY (monster_id) REFERENCES monsters (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- An action with a damage expression is an attack - one without is a trait the
-- DM reads out, which is why damage is allowed to be empty.
CREATE TABLE monster_actions (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    monster_id   BIGINT UNSIGNED NOT NULL,
    name         VARCHAR(120)    NOT NULL,
    attack_bonus INT             NOT NULL DEFAULT 0,
    damage       VARCHAR(40)     NOT NULL DEFAULT '',
    damage_type  VARCHAR(24)     NOT NULL DEFAULT '',
    description  TEXT            NOT NULL,
    PRIMARY KEY (id),
    KEY idx_monster_actions_monster (monster_id),
    CONSTRAINT fk_monster_actions_monster FOREIGN KEY (monster_id) REFERENCES monsters (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE tokens
    ADD COLUMN monster_id BIGINT UNSIGNED NULL AFTER character_id,
    ADD KEY idx_tokens_monster (monster_id),
    ADD CONSTRAINT fk_tokens_monster FOREIGN KEY (monster_id) REFERENCES monsters (id) ON DELETE SET NULL;
