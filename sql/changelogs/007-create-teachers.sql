--liquibase formatted sql

--changeset vidhya:007-create-teachers
CREATE TABLE IF NOT EXISTS teachers (
    user_id         VARCHAR(64) PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    school_id       VARCHAR(64) NOT NULL REFERENCES schools (id) ON DELETE CASCADE,
    employee_number VARCHAR(50) DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--changeset vidhya:007b-teachers-school-idx
CREATE INDEX IF NOT EXISTS ix_teachers_school_id ON teachers (school_id);
