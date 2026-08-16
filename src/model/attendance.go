package model

import "time"

// AttendanceStatus enumerates the values allowed for an attendance record.
type AttendanceStatus string

// Supported attendance statuses.
const (
	AttendancePresent AttendanceStatus = "PRESENT"
	AttendanceAbsent  AttendanceStatus = "ABSENT"
	AttendanceLate    AttendanceStatus = "LATE"
	AttendanceHalfDay AttendanceStatus = "HALF_DAY"
)

// IsValid reports whether s is one of the known statuses.
func (s AttendanceStatus) IsValid() bool {
	switch s {
	case AttendancePresent, AttendanceAbsent, AttendanceLate, AttendanceHalfDay:
		return true
	default:
		return false
	}
}

// Attendance records a single day's attendance for a student. SubjectID is
// optional: leave nil for whole-day attendance, or set it to record
// per-period/per-subject attendance.
type Attendance struct {
	ID        string           `json:"id" gorm:"column:id;primaryKey"`
	SchoolID  string           `json:"schoolId" gorm:"column:school_id"`
	StudentID string           `json:"studentId" gorm:"column:student_id"`
	ClassID   string           `json:"classId" gorm:"column:class_id"`
	Date      time.Time        `json:"date" gorm:"column:date"`
	Status    AttendanceStatus `json:"status" gorm:"column:status"`
	SubjectID *string          `json:"subjectId,omitempty" gorm:"column:subject_id"`
	MarkedBy  string           `json:"markedBy" gorm:"column:marked_by"`
	Remarks   string           `json:"remarks,omitempty" gorm:"column:remarks"`
	CreatedAt time.Time        `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time        `json:"updatedAt" gorm:"column:updated_at"`
}

// TableName pins the GORM table name explicitly.
func (Attendance) TableName() string { return "attendance" }

// AttendanceSummary aggregates a student's attendance over a date range.
type AttendanceSummary struct {
	StudentID       string  `json:"studentId"`
	TotalDays       int     `json:"totalDays"`
	PresentDays     int     `json:"presentDays"`
	AbsentDays      int     `json:"absentDays"`
	LateDays        int     `json:"lateDays"`
	HalfDays        int     `json:"halfDays"`
	AttendancePct   float64 `json:"attendancePercentage"`
}
