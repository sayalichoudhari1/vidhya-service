--liquibase formatted sql

--changeset vidhya:008-create-attendance
CREATE TABLE IF NOT EXISTS attendance (
    id         VARCHAR(64)  PRIMARY KEY,
    school_id  VARCHAR(64)  NOT NULL REFERENCES schools (id) ON DELETE CASCADE,
    student_id VARCHAR(64)  NOT NULL REFERENCES students (user_id) ON DELETE CASCADE,
    class_id   VARCHAR(64)  NOT NULL REFERENCES classes (id) ON DELETE CASCADE,
    date       DATE         NOT NULL,
    status     VARCHAR(20)  NOT NULL CHECK (status IN ('PRESENT', 'ABSENT', 'LATE', 'HALF_DAY')),
    subject_id VARCHAR(64)  REFERENCES subjects (id) ON DELETE SET NULL, -- NULL = whole-day attendance
    marked_by  VARCHAR(64)  NOT NULL REFERENCES users (id) ON DELETE SET NULL,
    remarks    VARCHAR(500) DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

--changeset vidhya:008b-attendance-unique-entry
-- One attendance record per student per day per subject (subject_id NULL
-- covers whole-day attendance); COALESCE keeps NULLs unique-comparable.
CREATE UNIQUE INDEX IF NOT EXISTS ux_attendance_student_date_subject
    ON attendance (student_id, date, COALESCE(subject_id, ''));

--changeset vidhya:008c-attendance-class-date-idx
CREATE INDEX IF NOT EXISTS ix_attendance_class_date ON attendance (class_id, date);

--changeset vidhya:008d-attendance-student-idx
CREATE INDEX IF NOT EXISTS ix_attendance_student_id ON attendance (student_id);
