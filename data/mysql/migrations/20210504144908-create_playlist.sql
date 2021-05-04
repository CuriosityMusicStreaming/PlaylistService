-- +migrate Up
CREATE TABLE playlist
(
    `playlist_id` binary(16) NOT NULL,
    `name` varchar(255) NOT NULL,
    `owner_id` binary(16) NOT NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    PRIMARY KEY (`playlist_id`),
    INDEX `playlist_id_index` (`playlist_id`),
    INDEX `owner_id_index` (`owner_id`)
);

-- +migrate Down
DROP TABLE playlist;