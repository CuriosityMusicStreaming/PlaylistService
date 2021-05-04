-- +migrate Up
CREATE TABLE playlist_item
(
    `playlist_item_id` binary(16) NOT NULL,
    `playlist_id` binary(16) NOT NULL,
    `content_id` binary(16) NOT NULL,
    `created_at` timestamp NOT NULL,
    PRIMARY KEY (`playlist_item_id`),
    INDEX `playlist_item_id_index` (`playlist_item_id`),
    FOREIGN KEY (`playlist_id`) REFERENCES playlist (`playlist_id`),
    INDEX `content_id_index` (`content_id`)
);
-- +migrate Down
DROP TABLE playlist_item;