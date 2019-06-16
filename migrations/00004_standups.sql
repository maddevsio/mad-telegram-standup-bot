-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standups` MODIFY `text` TEXT COLLATE utf8mb4_unicode_ci NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standups` MODIFY `text` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL;
