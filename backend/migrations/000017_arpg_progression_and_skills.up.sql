-- 000017_arpg_progression_and_skills.up.sql

-- 1. Characters: ARPG stats, XP, DM level gating
ALTER TABLE characters
    ADD COLUMN xp BIGINT NOT NULL DEFAULT 0 AFTER level,
    ADD COLUMN stat_points INT NOT NULL DEFAULT 0 AFTER xp,
    ADD COLUMN skill_points INT NOT NULL DEFAULT 0 AFTER stat_points,
    ADD COLUMN level_blocked TINYINT(1) NOT NULL DEFAULT 0 AFTER skill_points,
    ADD COLUMN block_reason VARCHAR(255) NOT NULL DEFAULT '' AFTER level_blocked;

-- 2. Items requirement and scroll classification
ALTER TABLE compendium_items
    ADD COLUMN req_level INT NOT NULL DEFAULT 1 AFTER description,
    ADD COLUMN req_class VARCHAR(40) NOT NULL DEFAULT '' AFTER req_level,
    ADD COLUMN req_str INT NOT NULL DEFAULT 0 AFTER req_class,
    ADD COLUMN req_dex INT NOT NULL DEFAULT 0 AFTER req_str,
    ADD COLUMN req_con INT NOT NULL DEFAULT 0 AFTER req_dex,
    ADD COLUMN req_int INT NOT NULL DEFAULT 0 AFTER req_con,
    ADD COLUMN req_wis INT NOT NULL DEFAULT 0 AFTER req_int,
    ADD COLUMN req_cha INT NOT NULL DEFAULT 0 AFTER req_wis,
    ADD COLUMN teaches_skill_id BIGINT UNSIGNED NULL AFTER req_cha;

ALTER TABLE character_items
    ADD COLUMN req_level INT NOT NULL DEFAULT 1 AFTER description,
    ADD COLUMN req_class VARCHAR(40) NOT NULL DEFAULT '' AFTER req_level,
    ADD COLUMN req_str INT NOT NULL DEFAULT 0 AFTER req_class,
    ADD COLUMN req_dex INT NOT NULL DEFAULT 0 AFTER req_str,
    ADD COLUMN req_con INT NOT NULL DEFAULT 0 AFTER req_dex,
    ADD COLUMN req_int INT NOT NULL DEFAULT 0 AFTER req_con,
    ADD COLUMN req_wis INT NOT NULL DEFAULT 0 AFTER req_int,
    ADD COLUMN req_cha INT NOT NULL DEFAULT 0 AFTER req_wis,
    ADD COLUMN teaches_skill_id BIGINT UNSIGNED NULL AFTER req_cha;

-- 3. Class Skill Tree Definitions
CREATE TABLE class_skills (
    id                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    class_name         VARCHAR(40)     NOT NULL, -- 'Mage', 'Assassin', 'Warrior', 'Archer', 'Cleric'
    name               VARCHAR(120)    NOT NULL,
    tier               INT             NOT NULL DEFAULT 1, -- 1 to 4 (4 = Ultimate)
    req_level          INT             NOT NULL DEFAULT 1,
    parent_skill_id    BIGINT UNSIGNED NULL,
    damage             VARCHAR(40)     NOT NULL DEFAULT '',
    damage_type        VARCHAR(24)     NOT NULL DEFAULT '',
    cooldown_rounds    INT             NOT NULL DEFAULT 0,
    cost_resource      VARCHAR(60)     NOT NULL DEFAULT '', -- e.g. '1 Action', '10 Mana / Slot'
    description        TEXT            NOT NULL,
    created_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_class_skills_class (class_name, tier),
    CONSTRAINT fk_class_skills_parent FOREIGN KEY (parent_skill_id) REFERENCES class_skills (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Character Learned Skills (Active loadout max 10)
CREATE TABLE character_skills (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    character_id BIGINT UNSIGNED NOT NULL,
    skill_id     BIGINT UNSIGNED NOT NULL,
    is_active    TINYINT(1)      NOT NULL DEFAULT 1, -- Combat deck (max 10 active)
    learned_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_character_skill (character_id, skill_id),
    KEY idx_character_skills_active (character_id, is_active),
    CONSTRAINT fk_character_skills_char FOREIGN KEY (character_id) REFERENCES characters (id) ON DELETE CASCADE,
    CONSTRAINT fk_character_skills_skill FOREIGN KEY (skill_id) REFERENCES class_skills (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. Monster Pending Drops & DM Review Queue
CREATE TABLE pending_loot_drops (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    campaign_id  BIGINT UNSIGNED NOT NULL,
    map_id       BIGINT UNSIGNED NOT NULL,
    monster_name VARCHAR(120)    NOT NULL,
    item_name    VARCHAR(120)    NOT NULL,
    item_id      BIGINT UNSIGNED NULL,
    quantity     INT             NOT NULL DEFAULT 1,
    cp           INT             NOT NULL DEFAULT 0,
    sp           INT             NOT NULL DEFAULT 0,
    gp           INT             NOT NULL DEFAULT 0,
    status       ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending',
    reject_reason VARCHAR(255)   NOT NULL DEFAULT '',
    assigned_to_character_id BIGINT UNSIGNED NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_pending_loot_campaign (campaign_id, status),
    CONSTRAINT fk_pending_loot_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE CASCADE,
    CONSTRAINT fk_pending_loot_map FOREIGN KEY (map_id) REFERENCES maps (id) ON DELETE CASCADE,
    CONSTRAINT fk_pending_loot_char FOREIGN KEY (assigned_to_character_id) REFERENCES characters (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 6. Pre-seed Default 5 Class Skill Trees (4 Skills each: T1, T2, T3, T4 Ult)
-- Mage
INSERT INTO class_skills (id, class_name, name, tier, req_level, parent_skill_id, damage, damage_type, cooldown_rounds, cost_resource, description) VALUES
(1, 'Mage', 'Arcane Missiles', 1, 1, NULL, '3d4+3', 'force', 0, '1 Action', 'Launches glowing darts of magical force that unfailingly strike targets.'),
(2, 'Mage', 'Frost Nova', 2, 4, 1, '2d8', 'cold', 2, '1 Action', 'Erupts freezing shards around you. Traps enemies and halves movement for 1 turn.'),
(3, 'Mage', 'Meteor Strike', 3, 10, 2, '8d6', 'fire', 4, '1 Action', 'Calls down a flaming celestial rock in a 20ft radius, burning ground.'),
(4, 'Mage', 'Time Warp', 4, 20, 3, '', '', 6, '1 Action', 'Bends the flow of time, granting user an immediate extra turn and haste.');

-- Assassin / Rogue
INSERT INTO class_skills (id, class_name, name, tier, req_level, parent_skill_id, damage, damage_type, cooldown_rounds, cost_resource, description) VALUES
(5, 'Assassin', 'Shadow Step', 1, 1, NULL, '', '', 1, 'Bonus Action', 'Steps through shadows up to 30ft behind target. Next attack rolls with Advantage.'),
(6, 'Assassin', 'Poison Strike', 2, 4, 5, '2d6+3', 'poison', 1, '1 Action', 'Envenomed blade strike dealing piercing damage plus poison over 2 turns.'),
(7, 'Assassin', 'Smoke Screen', 3, 10, 6, '', '', 3, '1 Action', 'Drops dense smoke in 15ft radius. Blinds opponents and grants free Disengage.'),
(8, 'Assassin', 'Death Mark', 4, 20, 7, '10d6', 'necrotic', 5, '1 Action', 'Executes target with precision. Automatically scores a Critical Hit if target is below 50% HP.');

-- Warrior
INSERT INTO class_skills (id, class_name, name, tier, req_level, parent_skill_id, damage, damage_type, cooldown_rounds, cost_resource, description) VALUES
(9, 'Warrior', 'Shield Bash', 1, 1, NULL, '1d6+STR', 'bludgeoning', 1, 'Bonus Action', 'Slams enemy with shield. Target must pass STR save or fall Prone.'),
(10, 'Warrior', 'Whirlwind', 2, 4, 9, '2d8+STR', 'slashing', 2, '1 Action', 'Spins weapon dealing slashing damage to all adjacent targets in melee reach.'),
(11, 'Warrior', 'Battle Cry', 3, 10, 10, '', '', 4, 'Bonus Action', 'Roars rallying allies, granting +15 Temporary HP and Advantage on next attack roll.'),
(12, 'Warrior', 'Avatar of War', 4, 20, 11, '', '', 6, '1 Action', 'Channels the god of battle. Grants Resistance to all damage and doubles weapon damage dice for 3 turns.');

-- Archer
INSERT INTO class_skills (id, class_name, name, tier, req_level, parent_skill_id, damage, damage_type, cooldown_rounds, cost_resource, description) VALUES
(13, 'Archer', 'Power Shot', 1, 1, NULL, '1d10+DEX', 'piercing', 0, '1 Action', 'Fires a high-velocity arrow pushing target back 10ft on hit.'),
(14, 'Archer', 'Rain of Arrows', 2, 4, 13, '3d8', 'piercing', 2, '1 Action', 'Showers a 15ft targeted zone with falling arrows hitting all enemies within.'),
(15, 'Archer', 'Eagle Eye', 3, 10, 14, '', '', 3, 'Bonus Action', 'Focuses vision. Critical hit threshold becomes 19-20 and weapon range doubles for 2 turns.'),
(16, 'Archer', 'Ballista Pierce', 4, 20, 15, '6d10', 'piercing', 5, '1 Action', 'Fires a devastating spear-like bolt in a 60ft straight line piercing all enemies.');

-- Cleric
INSERT INTO class_skills (id, class_name, name, tier, req_level, parent_skill_id, damage, damage_type, cooldown_rounds, cost_resource, description) VALUES
(17, 'Cleric', 'Healing Grace', 1, 1, NULL, '1d8+WIS', 'radiant', 0, 'Bonus Action', 'Restores hit points to self or chosen ally within 30ft.'),
(18, 'Cleric', 'Sacred Flame', 2, 4, 17, '2d8', 'radiant', 1, '1 Action', 'Flame-like radiance descends on a target, ignoring cover penalties.'),
(19, 'Cleric', 'Guardian Aegis', 3, 10, 18, '', '', 3, '1 Action', 'Manifests holy shields around allies in 20ft aura, adding +2 AC and curing blind/frightened.'),
(20, 'Cleric', 'Divine Miracle', 4, 20, 19, '4d8+20', 'radiant', 7, '1 Action', 'Channels divine power to mass heal all party members and revive fallen allies at 1 HP.');
