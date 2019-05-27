-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE `groups` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `chat_id` BIGINT NOT NULL,
    `title` VARCHAR(255) NOT NULL,
    `description` VARCHAR(255) NOT NULL,
    `standup_deadline` VARCHAR(255) NOT NULL, 
    `username` VARCHAR(255) NOT NULL,
    `tz` VARCHAR(255) NOT NULL
);
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE `groups`;