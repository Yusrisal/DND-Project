DROP TABLE combatants;
DROP TABLE encounters;
DROP TABLE character_proficiencies;
DROP TABLE token_conditions;

ALTER TABLE character_items
    DROP COLUMN damage_type,
    DROP COLUMN damage,
    DROP COLUMN attack_bonus;

ALTER TABLE tokens
    DROP FOREIGN KEY fk_tokens_character,
    DROP KEY idx_tokens_character,
    DROP COLUMN armor_class,
    DROP COLUMN hp_max,
    DROP COLUMN hp_current,
    DROP COLUMN character_id;
