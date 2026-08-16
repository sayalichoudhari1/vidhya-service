--liquibase formatted sql

--changeset vidhya:002-create-users
CREATE TABLE IF NOT EXISTS users (
    id            VARCHAR(64)  PRIMARY KEY,
    school_id     VARCHAR(64)  REFERENCES schools (id) ON DELETE SET NULL, -- NULL for platform ADMIN
    role          VARCHAR(20)  NOT NULL CHECK (role IN ('ADMIN', 'PRINCIPAL', 'TEACHER', 'STUDENT', 'PARENT')),
    first_name    VARCHAR(150) NOT NULL,
    last_name     VARCHAR(150) DEFAULT '',
    email         VARCHAR(255) NOT NULL,
    phone         VARCHAR(30)  DEFAULT '',
    password_hash VARCHAR(255) NOT NULL,
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

--changeset vidhya:002b-users-unique-email
-- Email must be unique among non-deleted users; a partial index allows the
-- same email to be reused after a soft-delete.
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email_active ON users (email) WHERE deleted_at IS NULL;

--changeset vidhya:002c-users-school-idx
CREATE INDEX IF NOT EXISTS ix_users_school_id ON users (school_id);

--changeset vidhya:002d-schools-principal-fk
-- Added after users exists to avoid a circular creation-order dependency.
ALTER TABLE schools
    ADD CONSTRAINT fk_schools_principal
    FOREIGN KEY (principal_id) REFERENCES users (id) ON DELETE SET NULL;
