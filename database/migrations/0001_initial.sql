-- +migrate Up
CREATE TABLE version (
    image VARCHAR PRIMARY KEY,
    version VARCHAR NOT NULL,
    CONSTRAINT constraint_image UNIQUE (image)
);

-- +migrate Down
DROP TABLE version;
