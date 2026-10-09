ALTER TABLE token_conditions
    DROP COLUMN duration_rounds,
    DROP COLUMN remaining_rounds,
    DROP COLUMN notes;

DROP TABLE IF EXISTS status_effects;
