-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `groups` ADD `onbording_message` TEXT COLLATE utf8mb4_unicode_ci NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `groups` DROP `onbording_message` TEXT COLLATE utf8mb4_unicode_ci NOT NULL;
