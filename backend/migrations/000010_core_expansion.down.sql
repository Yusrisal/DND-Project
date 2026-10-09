DROP TABLE IF EXISTS shop_items;
DROP TABLE IF EXISTS shops;

ALTER TABLE monsters
    DROP COLUMN damage_resistances,
    DROP COLUMN damage_vulnerabilities,
    DROP COLUMN damage_immunities,
    DROP COLUMN condition_immunities;

ALTER TABLE character_items
    DROP COLUMN cost_cp;

ALTER TABLE characters
    DROP COLUMN status,
    DROP COLUMN cp,
    DROP COLUMN sp,
    DROP COLUMN ep,
    DROP COLUMN gp,
    DROP COLUMN pp,
    DROP COLUMN death_save_successes,
    DROP COLUMN death_save_failures,
    DROP COLUMN hit_dice_total,
    DROP COLUMN hit_dice_used,
    DROP COLUMN hit_dice_die,
    DROP COLUMN damage_resistances,
    DROP COLUMN damage_vulnerabilities,
    DROP COLUMN damage_immunities,
    DROP COLUMN condition_immunities;
