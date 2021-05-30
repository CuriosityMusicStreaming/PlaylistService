-- +migrate Up
CREATE TABLE tracked_stored_event
(
    `transport_name` VARCHAR(255) NOT NULL ,
    `last_stored_event_id` binary(16) NOT NULL,
    `created_at` datetime NOT NULL,
    PRIMARY KEY (`transport_name`)
);
-- +migrate Down
DROP TABLE tracked_stored_event;