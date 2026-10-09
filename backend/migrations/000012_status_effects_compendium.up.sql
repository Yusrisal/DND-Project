-- Status Effects Compendium (standard conditions + custom buffs/debuffs)
CREATE TABLE status_effects (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    campaign_id             BIGINT UNSIGNED NULL,
    name                    VARCHAR(120)    NOT NULL,
    kind                    ENUM('condition', 'buff', 'debuff', 'hazard') NOT NULL DEFAULT 'condition',
    default_duration_rounds INT             NOT NULL DEFAULT 0,
    color                   VARCHAR(16)     NOT NULL DEFAULT '#e67e22',
    effect_description      TEXT            NOT NULL,
    created_at              DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_status_effects_campaign (campaign_id),
    KEY idx_status_effects_name (name),
    CONSTRAINT fk_status_effects_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Enhance token conditions with duration tracking and notes
ALTER TABLE token_conditions
    ADD COLUMN duration_rounds INT NOT NULL DEFAULT 0 AFTER condition_name,
    ADD COLUMN remaining_rounds INT NOT NULL DEFAULT 0 AFTER duration_rounds,
    ADD COLUMN notes VARCHAR(255) NOT NULL DEFAULT '' AFTER remaining_rounds;
