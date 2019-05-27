-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE `standups` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `message_id` INTEGER NOT NULL,
    `created` DATETIME NOT NULL,
    `modified` DATETIME NOT NULL,
    `username` VARCHAR(255) NOT NULL,
    `text` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    `chat_id` BIGINT NOT NULL,
    KEY (`created`, `username`)
);
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE `standups`;