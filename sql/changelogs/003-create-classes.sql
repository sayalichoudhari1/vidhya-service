--liquibase formatted sql

--changeset vidhya:003-create-classes
CREATE TABLE IF NOT EXISTS classes (
    id               VARCHAR(64)  PRIMARY KEY,
    school_id        VARCHAR(64)  NOT NULL REFERENCES schools (id) ON DELETE CASCADE,
    name             VARCHAR(50)  NOT NULL,
    section          VARCHAR(20)  DEFAULT '',
    academic_year    VARCHAR(20)  NOT NULL,
    class_teacher_id VARCHAR(64)  REFERENCES users (id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

--changeset vidhya:003b-classes-school-idx
CREATE INDEX IF NOT EXISTS ix_classes_school_id ON classes (school_id);

--changeset vidhya:003c-classes-unique-per-school
CREATE UNIQUE INDEX IF NOT EXISTS ux_classes_school_name_section_year
    ON classes (school_id, name, section, academic_year);
