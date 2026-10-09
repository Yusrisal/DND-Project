ALTER TABLE tokens
    DROP FOREIGN KEY fk_tokens_monster,
    DROP KEY idx_tokens_monster,
    DROP COLUMN monster_id;

DROP TABLE monster_actions;
DROP TABLE monster_proficiencies;
DROP TABLE monsters;
