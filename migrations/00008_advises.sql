-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `groups` ADD `advises` VARCHAR(255) NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `groups` DROP `advises` VARCHAR(255) NOT NULL;
