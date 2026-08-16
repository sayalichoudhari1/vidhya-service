--liquibase formatted sql

--changeset vidhya:010-create-exam-results
CREATE TABLE IF NOT EXISTS exam_results (
    id             VARCHAR(64)  PRIMARY KEY,
    exam_id        VARCHAR(64)  NOT NULL REFERENCES exams (id) ON DELETE CASCADE,
    student_id     VARCHAR(64)  NOT NULL REFERENCES students (user_id) ON DELETE CASCADE,
    marks_obtained NUMERIC(6,2) NOT NULL,
    grade          VARCHAR(5)   DEFAULT '',
    remarks        VARCHAR(500) DEFAULT '',
    evaluated_by   VARCHAR(64)  NOT NULL REFERENCES users (id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

--changeset vidhya:010b-exam-results-unique
CREATE UNIQUE INDEX IF NOT EXISTS ux_exam_results_exam_student ON exam_results (exam_id, student_id);

--changeset vidhya:010c-exam-results-student-idx
CREATE INDEX IF NOT EXISTS ix_exam_results_student_id ON exam_results (student_id);
