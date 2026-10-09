ALTER TABLE tokens
    ADD COLUMN owner_user_id BIGINT UNSIGNED NULL AFTER map_id,
    ADD KEY idx_tokens_owner_user_id (owner_user_id),
    ADD CONSTRAINT fk_tokens_owner FOREIGN KEY (owner_user_id) REFERENCES users (id) ON DELETE SET NULL;
