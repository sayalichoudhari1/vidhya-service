--liquibase formatted sql

--changeset vidhya:009-create-exams
CREATE TABLE IF NOT EXISTS exams (
    id         VARCHAR(64)   PRIMARY KEY,
    school_id  VARCHAR(64)   NOT NULL REFERENCES schools (id) ON DELETE CASCADE,
    name       VARCHAR(255)  NOT NULL,
    class_id   VARCHAR(64)   NOT NULL REFERENCES classes (id) ON DELETE CASCADE,
    subject_id VARCHAR(64)   NOT NULL REFERENCES subjects (id) ON DELETE CASCADE,
    exam_type  VARCHAR(20)   NOT NULL CHECK (exam_type IN ('UNIT_TEST', 'MID_TERM', 'FINAL', 'QUIZ')),
    exam_date  DATE          NOT NULL,
    max_marks  NUMERIC(6,2)  NOT NULL,
    created_by VARCHAR(64)   NOT NULL REFERENCES users (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

--changeset vidhya:009b-exams-class-idx
CREATE INDEX IF NOT EXISTS ix_exams_class_id ON exams (class_id);

--changeset vidhya:009c-exams-school-idx
CREATE INDEX IF NOT EXISTS ix_exams_school_id ON exams (school_id);
