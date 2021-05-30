-- +migrate Up
CREATE TABLE stored_event
(
    `stored_event_id` binary(16) NOT NULL,
    `type` VARCHAR(255) NOT NULL,
    `body` VARCHAR(1000) NOT NULL,
    `created_at` datetime NOT NULL,
    PRIMARY KEY (`stored_event_id`)
);
-- +migrate Down
DROP TABLE stored_event;