--liquibase formatted sql

--changeset vidhya:001-create-schools
CREATE TABLE IF NOT EXISTS schools (
    id           VARCHAR(64)  PRIMARY KEY,
    name         VARCHAR(255) NOT NULL,
    code         VARCHAR(50)  NOT NULL,
    address      VARCHAR(500) DEFAULT '',
    principal_id VARCHAR(64),
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

--changeset vidhya:001b-schools-unique-code
CREATE UNIQUE INDEX IF NOT EXISTS ux_schools_code ON schools (code);
