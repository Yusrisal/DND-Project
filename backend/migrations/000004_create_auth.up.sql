CREATE TABLE users (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    email         VARCHAR(191)    NOT NULL,
    display_name  VARCHAR(80)     NOT NULL,
    password_hash VARCHAR(255)    NOT NULL,
    role          ENUM('dm','player') NOT NULL DEFAULT 'player',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE sessions (
    token_hash CHAR(64)        NOT NULL,
    user_id    BIGINT UNSIGNED NOT NULL,
    expires_at DATETIME        NOT NULL,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (token_hash),
    KEY idx_sessions_user_id (user_id),
    KEY idx_sessions_expires_at (expires_at),
    CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE invites (
    code            VARCHAR(32)     NOT NULL,
    role            ENUM('dm','player') NOT NULL DEFAULT 'player',
    created_by      BIGINT UNSIGNED NOT NULL,
    used_by         BIGINT UNSIGNED NULL,
    used_at         DATETIME        NULL,
    expires_at      DATETIME        NOT NULL,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (code),
    KEY idx_invites_created_by (created_by),
    CONSTRAINT fk_invites_creator FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_invites_user FOREIGN KEY (used_by) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
