--liquibase formatted sql

--changeset vidhya:005-create-class-subjects
CREATE TABLE IF NOT EXISTS class_subjects (
    id         VARCHAR(64) PRIMARY KEY,
    class_id   VARCHAR(64) NOT NULL REFERENCES classes (id) ON DELETE CASCADE,
    subject_id VARCHAR(64) NOT NULL REFERENCES subjects (id) ON DELETE CASCADE,
    teacher_id VARCHAR(64) REFERENCES users (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--changeset vidhya:005b-class-subjects-unique
CREATE UNIQUE INDEX IF NOT EXISTS ux_class_subjects_class_subject ON class_subjects (class_id, subject_id);

--changeset vidhya:005c-class-subjects-idx
CREATE INDEX IF NOT EXISTS ix_class_subjects_teacher_id ON class_subjects (teacher_id);
