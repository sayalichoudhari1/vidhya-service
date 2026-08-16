package model

import "time"

// Class represents a grade/section within a school, e.g. "10-A" for the
// 2026-27 academic year.
type Class struct {
	ID             string    `json:"id" gorm:"column:id;primaryKey"`
	SchoolID       string    `json:"schoolId" gorm:"column:school_id"`
	Name           string    `json:"name" gorm:"column:name"`       // e.g. "10"
	Section        string    `json:"section" gorm:"column:section"` // e.g. "A"
	AcademicYear   string    `json:"academicYear" gorm:"column:academic_year"`
	ClassTeacherID *string   `json:"classTeacherId,omitempty" gorm:"column:class_teacher_id"`
	CreatedAt      time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// TableName pins the GORM table name explicitly.
func (Class) TableName() string { return "classes" }

// Subject represents a teachable subject, e.g. "Mathematics".
type Subject struct {
	ID        string    `json:"id" gorm:"column:id;primaryKey"`
	SchoolID  string    `json:"schoolId" gorm:"column:school_id"`
	Name      string    `json:"name" gorm:"column:name"`
	Code      string    `json:"code,omitempty" gorm:"column:code"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// TableName pins the GORM table name explicitly.
func (Subject) TableName() string { return "subjects" }

// ClassSubject links a subject to a class and assigns the teacher who
// teaches it, e.g. "10-A takes Mathematics, taught by Mr. Sharma".
type ClassSubject struct {
	ID        string    `json:"id" gorm:"column:id;primaryKey"`
	ClassID   string    `json:"classId" gorm:"column:class_id"`
	SubjectID string    `json:"subjectId" gorm:"column:subject_id"`
	TeacherID *string   `json:"teacherId,omitempty" gorm:"column:teacher_id"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
}

// TableName pins the GORM table name explicitly.
func (ClassSubject) TableName() string { return "class_subjects" }
