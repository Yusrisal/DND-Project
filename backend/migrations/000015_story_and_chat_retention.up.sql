-- Add chat & story retention window in days (0 = keep all visible, 1 = 24h, 7 = 1 week, 14 = 2 weeks, 30 = 1 month)
ALTER TABLE campaigns
    ADD COLUMN chat_retention_days INT NOT NULL DEFAULT 0 AFTER pause_message;
