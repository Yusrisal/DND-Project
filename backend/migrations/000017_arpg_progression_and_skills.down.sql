-- 000017_arpg_progression_and_skills.down.sql

DROP TABLE IF EXISTS pending_loot_drops;
DROP TABLE IF EXISTS character_skills;
DROP TABLE IF EXISTS class_skills;

ALTER TABLE character_items
    DROP COLUMN IF EXISTS teaches_skill_id,
    DROP COLUMN IF EXISTS req_cha,
    DROP COLUMN IF EXISTS req_wis,
    DROP COLUMN IF EXISTS req_int,
    DROP COLUMN IF EXISTS req_con,
    DROP COLUMN IF EXISTS req_dex,
    DROP COLUMN IF EXISTS req_str,
    DROP COLUMN IF EXISTS req_class,
    DROP COLUMN IF EXISTS req_level;

ALTER TABLE compendium_items
    DROP COLUMN IF EXISTS teaches_skill_id,
    DROP COLUMN IF EXISTS req_cha,
    DROP COLUMN IF EXISTS req_wis,
    DROP COLUMN IF EXISTS req_int,
    DROP COLUMN IF EXISTS req_con,
    DROP COLUMN IF EXISTS req_dex,
    DROP COLUMN IF EXISTS req_str,
    DROP COLUMN IF EXISTS req_class,
    DROP COLUMN IF EXISTS req_level;

ALTER TABLE characters
    DROP COLUMN IF EXISTS block_reason,
    DROP COLUMN IF EXISTS level_blocked,
    DROP COLUMN IF EXISTS skill_points,
    DROP COLUMN IF EXISTS stat_points,
    DROP COLUMN IF EXISTS xp;
