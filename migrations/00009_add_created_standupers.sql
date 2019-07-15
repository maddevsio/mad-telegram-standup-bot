-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standupers` ADD `created` DATETIME NOT NULL DEFAULT '2019-01-01 00:00:00';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standupers` DROP `created`;
