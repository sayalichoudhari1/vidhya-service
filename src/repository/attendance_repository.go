package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"vidhya-service/src/model"
)

// AttendanceRepository persists daily/per-subject attendance records.
type AttendanceRepository struct {
	db *gorm.DB
}

// NewAttendanceRepository creates an AttendanceRepository.
func NewAttendanceRepository(db *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

// Create inserts a new attendance record.
func (r *AttendanceRepository) Create(ctx context.Context, a *model.Attendance) error {
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return fmt.Errorf("AttendanceRepository.Create: %w", err)
	}
	return nil
}

// Upsert inserts an attendance record, or updates the status/remarks if one
// already exists for the same (student, date, subject) - this is what
// backs "mark attendance" being safely called more than once per day.
func (r *AttendanceRepository) Upsert(ctx context.Context, a *model.Attendance) error {
	var existing model.Attendance
	q := r.db.WithContext(ctx).Where("student_id = ? AND date = ?", a.StudentID, a.Date)
	if a.SubjectID != nil {
		q = q.Where("subject_id = ?", *a.SubjectID)
	} else {
		q = q.Where("subject_id IS NULL")
	}

	err := q.First(&existing).Error
	switch {
	case err == nil:
		existing.Status = a.Status
		existing.Remarks = a.Remarks
		existing.MarkedBy = a.MarkedBy
		if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return fmt.Errorf("AttendanceRepository.Upsert: update: %w", err)
		}
		*a = existing
		return nil
	case err == gorm.ErrRecordNotFound:
		if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
			return fmt.Errorf("AttendanceRepository.Upsert: create: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("AttendanceRepository.Upsert: lookup: %w", err)
	}
}

// GetByID fetches an attendance record by ID.
func (r *AttendanceRepository) GetByID(ctx context.Context, id string) (*model.Attendance, error) {
	var a model.Attendance
	if err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("AttendanceRepository.GetByID: %w", err)
	}
	return &a, nil
}

// ListForStudent lists a student's attendance records within [from, to].
func (r *AttendanceRepository) ListForStudent(ctx context.Context, studentID string, from, to time.Time) ([]model.Attendance, error) {
	var rows []model.Attendance
	err := r.db.WithContext(ctx).
		Where("student_id = ? AND date BETWEEN ? AND ?", studentID, from, to).
		Order("date").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepository.ListForStudent: %w", err)
	}
	return rows, nil
}

// ListForClassOnDate lists every student's attendance for a class on a given
// date - used by teachers to review/adjust a day's register.
func (r *AttendanceRepository) ListForClassOnDate(ctx context.Context, classID string, date time.Time) ([]model.Attendance, error) {
	var rows []model.Attendance
	if err := r.db.WithContext(ctx).Where("class_id = ? AND date = ?", classID, date).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("AttendanceRepository.ListForClassOnDate: %w", err)
	}
	return rows, nil
}

// Update persists changes to an existing attendance record.
func (r *AttendanceRepository) Update(ctx context.Context, a *model.Attendance) error {
	if err := r.db.WithContext(ctx).Model(&model.Attendance{}).Where("id = ?", a.ID).Updates(a).Error; err != nil {
		return fmt.Errorf("AttendanceRepository.Update: %w", err)
	}
	return nil
}

// Delete removes an attendance record by ID.
func (r *AttendanceRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Attendance{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("AttendanceRepository.Delete: %w", err)
	}
	return nil
}

// Summary computes the attendance summary for a student within [from, to].
func (r *AttendanceRepository) Summary(ctx context.Context, studentID string, from, to time.Time) (*model.AttendanceSummary, error) {
	rows, err := r.ListForStudent(ctx, studentID, from, to)
	if err != nil {
		return nil, err
	}

	summary := &model.AttendanceSummary{StudentID: studentID}
	for _, row := range rows {
		summary.TotalDays++
		switch row.Status {
		case model.AttendancePresent:
			summary.PresentDays++
		case model.AttendanceAbsent:
			summary.AbsentDays++
		case model.AttendanceLate:
			summary.LateDays++
		case model.AttendanceHalfDay:
			summary.HalfDays++
		}
	}
	if summary.TotalDays > 0 {
		summary.AttendancePct = float64(summary.PresentDays+summary.LateDays+summary.HalfDays) / float64(summary.TotalDays) * 100
	}
	return summary, nil
}
