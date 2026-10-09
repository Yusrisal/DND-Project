ALTER TABLE tokens
    DROP FOREIGN KEY fk_tokens_owner,
    DROP KEY idx_tokens_owner_user_id,
    DROP COLUMN owner_user_id;
