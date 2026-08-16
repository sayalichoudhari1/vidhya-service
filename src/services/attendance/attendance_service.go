// Package attendance implements marking and retrieving student attendance.
package attendance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"vidhya-service/src/apperrors"
	"vidhya-service/src/model"
	"vidhya-service/src/repository"
)

// MarkRequest is the payload for POST /attendance - marks (or corrects) one
// student's attendance for one day (and optionally one subject/period).
type MarkRequest struct {
	StudentID string  `json:"studentId"`
	ClassID   string  `json:"classId"`
	Date      string  `json:"date"` // YYYY-MM-DD
	Status    string  `json:"status"`
	SubjectID *string `json:"subjectId,omitempty"`
	Remarks   string  `json:"remarks,omitempty"`
}

// BulkMarkRequest marks attendance for an entire class at once - the common
// "take today's attendance" teacher workflow.
type BulkMarkRequest struct {
	ClassID  string        `json:"classId"`
	Date     string        `json:"date"`
	Entries  []BulkEntry   `json:"entries"`
	SubjectID *string      `json:"subjectId,omitempty"`
}

// BulkEntry is one student's status within a BulkMarkRequest.
type BulkEntry struct {
	StudentID string `json:"studentId"`
	Status    string `json:"status"`
	Remarks   string `json:"remarks,omitempty"`
}

// Service implements attendance marking/retrieval.
type Service struct {
	repo *repository.AttendanceRepository
}

// NewService creates an attendance Service.
func NewService(repo *repository.AttendanceRepository) *Service {
	return &Service{repo: repo}
}

// Mark records/updates a single attendance entry.
func (s *Service) Mark(ctx context.Context, schoolID, markedBy string, req MarkRequest) (*model.Attendance, error) {
	date, status, err := s.validate(req.Date, req.Status)
	if err != nil {
		return nil, err
	}
	if req.StudentID == "" || req.ClassID == "" {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "studentId and classId are required")
	}

	record := &model.Attendance{
		ID:        uuid.NewString(),
		SchoolID:  schoolID,
		StudentID: req.StudentID,
		ClassID:   req.ClassID,
		Date:      date,
		Status:    status,
		SubjectID: req.SubjectID,
		MarkedBy:  markedBy,
		Remarks:   req.Remarks,
	}
	if err := s.repo.Upsert(ctx, record); err != nil {
		return nil, fmt.Errorf("attendance.Mark: %w", err)
	}
	return record, nil
}

// MarkBulk records attendance for a whole class in one call.
func (s *Service) MarkBulk(ctx context.Context, schoolID, markedBy string, req BulkMarkRequest) ([]model.Attendance, error) {
	date, _, err := s.validate(req.Date, string(model.AttendancePresent))
	if err != nil {
		return nil, err
	}
	if req.ClassID == "" {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "classId is required")
	}
	if len(req.Entries) == 0 {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "entries must not be empty")
	}

	results := make([]model.Attendance, 0, len(req.Entries))
	for _, entry := range req.Entries {
		status := model.AttendanceStatus(entry.Status)
		if !status.IsValid() {
			return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, fmt.Sprintf("invalid status %q for student %q", entry.Status, entry.StudentID))
		}
		record := &model.Attendance{
			ID:        uuid.NewString(),
			SchoolID:  schoolID,
			StudentID: entry.StudentID,
			ClassID:   req.ClassID,
			Date:      date,
			Status:    status,
			SubjectID: req.SubjectID,
			MarkedBy:  markedBy,
			Remarks:   entry.Remarks,
		}
		if err := s.repo.Upsert(ctx, record); err != nil {
			return nil, fmt.Errorf("attendance.MarkBulk: student %s: %w", entry.StudentID, err)
		}
		results = append(results, *record)
	}
	return results, nil
}

// ListForStudent returns a student's attendance records within a date range.
// Defaults to the current month if from/to are empty.
func (s *Service) ListForStudent(ctx context.Context, studentID, fromStr, toStr string) ([]model.Attendance, error) {
	from, to, err := parseRange(fromStr, toStr)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListForStudent(ctx, studentID, from, to)
	if err != nil {
		return nil, fmt.Errorf("attendance.ListForStudent: %w", err)
	}
	return rows, nil
}

// Summary computes a student's attendance percentage within a date range.
func (s *Service) Summary(ctx context.Context, studentID, fromStr, toStr string) (*model.AttendanceSummary, error) {
	from, to, err := parseRange(fromStr, toStr)
	if err != nil {
		return nil, err
	}
	summary, err := s.repo.Summary(ctx, studentID, from, to)
	if err != nil {
		return nil, fmt.Errorf("attendance.Summary: %w", err)
	}
	return summary, nil
}

// ListForClassOnDate lists a class's full attendance register for one day.
func (s *Service) ListForClassOnDate(ctx context.Context, classID, dateStr string) ([]model.Attendance, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "date must be in YYYY-MM-DD format")
	}
	rows, err := s.repo.ListForClassOnDate(ctx, classID, date)
	if err != nil {
		return nil, fmt.Errorf("attendance.ListForClassOnDate: %w", err)
	}
	return rows, nil
}

// Delete removes an attendance record by ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("attendance.Delete: lookup: %w", err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("attendance.Delete: %w", err)
	}
	return nil
}

func (s *Service) validate(dateStr, statusStr string) (time.Time, model.AttendanceStatus, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, "", apperrors.WithMessage(apperrors.ErrInvalidPayload, "date must be in YYYY-MM-DD format")
	}
	status := model.AttendanceStatus(statusStr)
	if !status.IsValid() {
		return time.Time{}, "", apperrors.WithMessage(apperrors.ErrInvalidPayload, "status must be one of PRESENT, ABSENT, LATE, HALF_DAY")
	}
	return date, status, nil
}

func parseRange(fromStr, toStr string) (time.Time, time.Time, error) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := now

	if fromStr != "" {
		parsed, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, apperrors.WithMessage(apperrors.ErrInvalidPayload, "from must be in YYYY-MM-DD format")
		}
		from = parsed
	}
	if toStr != "" {
		parsed, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return time.Time{}, time.Time{}, apperrors.WithMessage(apperrors.ErrInvalidPayload, "to must be in YYYY-MM-DD format")
		}
		to = parsed
	}
	return from, to, nil
}
