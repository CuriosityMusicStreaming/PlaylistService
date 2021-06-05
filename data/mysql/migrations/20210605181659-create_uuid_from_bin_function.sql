-- +migrate Up
DELIMITER $$
DROP FUNCTION IF EXISTS uuid_from_bin$$
CREATE FUNCTION uuid_from_bin(uuid BINARY(16)) RETURNS VARCHAR(255)
BEGIN
    DECLARE result VARCHAR(255) DEFAULT "";
    SET result = LOWER(CONCAT(
            SUBSTR(HEX(uuid), 1, 8), '-',
            SUBSTR(HEX(uuid), 9, 4), '-',
            SUBSTR(HEX(uuid), 13, 4), '-',
            SUBSTR(HEX(uuid), 17, 4), '-',
            SUBSTR(HEX(uuid), 21)
        ));
    RETURN result;
END;$$
-- +migrate Down
DROP FUNCTION uuid_from_bin;