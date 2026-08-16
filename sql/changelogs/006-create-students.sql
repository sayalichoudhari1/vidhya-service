--liquibase formatted sql

--changeset vidhya:006-create-students
CREATE TABLE IF NOT EXISTS students (
    user_id          VARCHAR(64)  PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    school_id        VARCHAR(64)  NOT NULL REFERENCES schools (id) ON DELETE CASCADE,
    admission_number VARCHAR(50)  NOT NULL,
    roll_number      VARCHAR(20)  DEFAULT '',
    class_id         VARCHAR(64)  REFERENCES classes (id) ON DELETE SET NULL,
    parent_id        VARCHAR(64)  REFERENCES users (id) ON DELETE SET NULL,
    date_of_birth    DATE,
    gender           VARCHAR(20)  DEFAULT '',
    address          VARCHAR(500) DEFAULT '',
    blood_group      VARCHAR(10)  DEFAULT '',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

--changeset vidhya:006b-students-school-idx
CREATE INDEX IF NOT EXISTS ix_students_school_id ON students (school_id);

--changeset vidhya:006c-students-class-idx
CREATE INDEX IF NOT EXISTS ix_students_class_id ON students (class_id);

--changeset vidhya:006d-students-parent-idx
CREATE INDEX IF NOT EXISTS ix_students_parent_id ON students (parent_id);

--changeset vidhya:006e-students-unique-admission
CREATE UNIQUE INDEX IF NOT EXISTS ux_students_school_admission_number
    ON students (school_id, admission_number);
