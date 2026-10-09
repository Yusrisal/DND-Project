DROP TABLE IF EXISTS handouts;
DROP TABLE IF EXISTS aoe_templates;
ALTER TABLE maps
    DROP COLUMN fog_revealed,
    DROP COLUMN fog_enabled;
