--liquibase formatted sql

-- These constraints back the duplicate-detection checks the application
-- layer already runs (see src/services/*). They exist as a second line of
-- defense: even if a future code path forgets the application-level check,
-- or two concurrent requests race past it, Postgres still guarantees the
-- data cannot end up duplicated.

--changeset vidhya:011a-schools-unique-name-active
-- Two active schools cannot share a display name (their `code` was already
-- unique; this additionally catches "same school, different code" mistakes).
CREATE UNIQUE INDEX IF NOT EXISTS ux_schools_name_active
    ON schools (lower(name)) WHERE is_active = true;

--changeset vidhya:011b-users-unique-name-phone-active
-- Same person (first name + last name + phone), within the same school,
-- cannot be onboarded twice. COALESCE normalizes NULL school_id (platform
-- ADMIN accounts) to '' so the constraint still applies to them.
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_school_name_phone_active
    ON users (COALESCE(school_id, ''), lower(first_name), lower(last_name), phone)
    WHERE phone <> '' AND is_active = true;

--changeset vidhya:011c-teachers-unique-employee-number
-- A school cannot issue the same staff/employee number to two teachers.
CREATE UNIQUE INDEX IF NOT EXISTS ux_teachers_school_employee_number
    ON teachers (school_id, employee_number) WHERE employee_number <> '';

--changeset vidhya:011d-students-unique-class-roll-number
-- Two students in the same class cannot share a roll number.
CREATE UNIQUE INDEX IF NOT EXISTS ux_students_class_roll_number
    ON students (class_id, roll_number) WHERE roll_number <> '' AND class_id IS NOT NULL;

--changeset vidhya:011e-exams-unique-class-subject-type-date
-- The same exam (type + subject + class) cannot be scheduled twice on the
-- same date.
CREATE UNIQUE INDEX IF NOT EXISTS ux_exams_class_subject_type_date
    ON exams (class_id, subject_id, exam_type, exam_date);
