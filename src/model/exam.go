package model

import "time"

// ExamType enumerates common assessment types.
type ExamType string

// Supported exam types.
const (
	ExamTypeUnitTest ExamType = "UNIT_TEST"
	ExamTypeMidTerm  ExamType = "MID_TERM"
	ExamTypeFinal    ExamType = "FINAL"
	ExamTypeQuiz     ExamType = "QUIZ"
)

// Exam represents a single assessment for one subject in one class, e.g.
// "Mid-Term Mathematics, Class 10-A".
type Exam struct {
	ID        string    `json:"id" gorm:"column:id;primaryKey"`
	SchoolID  string    `json:"schoolId" gorm:"column:school_id"`
	Name      string    `json:"name" gorm:"column:name"`
	ClassID   string    `json:"classId" gorm:"column:class_id"`
	SubjectID string    `json:"subjectId" gorm:"column:subject_id"`
	ExamType  ExamType  `json:"examType" gorm:"column:exam_type"`
	ExamDate  time.Time `json:"examDate" gorm:"column:exam_date"`
	MaxMarks  float64   `json:"maxMarks" gorm:"column:max_marks"`
	CreatedBy string    `json:"createdBy" gorm:"column:created_by"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// TableName pins the GORM table name explicitly.
func (Exam) TableName() string { return "exams" }

// ExamResult records one student's marks for one exam. Grade is derived and
// stored for fast reads (see services/exam for the grading rule).
type ExamResult struct {
	ID             string    `json:"id" gorm:"column:id;primaryKey"`
	ExamID         string    `json:"examId" gorm:"column:exam_id"`
	StudentID      string    `json:"studentId" gorm:"column:student_id"`
	MarksObtained  float64   `json:"marksObtained" gorm:"column:marks_obtained"`
	Grade          string    `json:"grade,omitempty" gorm:"column:grade"`
	Remarks        string    `json:"remarks,omitempty" gorm:"column:remarks"`
	EvaluatedBy    string    `json:"evaluatedBy" gorm:"column:evaluated_by"`
	CreatedAt      time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// TableName pins the GORM table name explicitly.
func (ExamResult) TableName() string { return "exam_results" }

// MarksheetEntry is one row of a student's consolidated marksheet.
type MarksheetEntry struct {
	ExamID        string    `json:"examId"`
	ExamName      string    `json:"examName"`
	SubjectID     string    `json:"subjectId"`
	SubjectName   string    `json:"subjectName"`
	ExamType      ExamType  `json:"examType"`
	ExamDate      time.Time `json:"examDate"`
	MaxMarks      float64   `json:"maxMarks"`
	MarksObtained float64   `json:"marksObtained"`
	Grade         string    `json:"grade,omitempty"`
}

// Marksheet is a student's full consolidated result set, optionally filtered
// to a single class/academic year by the caller.
type Marksheet struct {
	StudentID string           `json:"studentId"`
	ClassID   string           `json:"classId,omitempty"`
	Entries   []MarksheetEntry `json:"entries"`
}
