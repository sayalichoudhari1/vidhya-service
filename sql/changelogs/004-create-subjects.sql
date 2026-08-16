--liquibase formatted sql

--changeset vidhya:004-create-subjects
CREATE TABLE IF NOT EXISTS subjects (
    id         VARCHAR(64)  PRIMARY KEY,
    school_id  VARCHAR(64)  NOT NULL REFERENCES schools (id) ON DELETE CASCADE,
    name       VARCHAR(150) NOT NULL,
    code       VARCHAR(50)  DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

--changeset vidhya:004b-subjects-school-idx
CREATE INDEX IF NOT EXISTS ix_subjects_school_id ON subjects (school_id);

--changeset vidhya:004c-subjects-unique-per-school
CREATE UNIQUE INDEX IF NOT EXISTS ux_subjects_school_name ON subjects (school_id, name);
