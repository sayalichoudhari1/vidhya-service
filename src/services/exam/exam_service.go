// Package exam implements exam creation and marks entry, and builds a
// student's consolidated marksheet.
package exam

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

// CreateExamRequest is the payload for POST /exams.
type CreateExamRequest struct {
	Name      string  `json:"name"`
	ClassID   string  `json:"classId"`
	SubjectID string  `json:"subjectId"`
	ExamType  string  `json:"examType"`
	ExamDate  string  `json:"examDate"` // YYYY-MM-DD
	MaxMarks  float64 `json:"maxMarks"`
}

// RecordMarksRequest is the payload for POST /exams/{id}/results - records
// one student's marks for that exam.
type RecordMarksRequest struct {
	StudentID     string  `json:"studentId"`
	MarksObtained float64 `json:"marksObtained"`
	Remarks       string  `json:"remarks,omitempty"`
}

// BulkRecordMarksRequest records marks for multiple students in one call -
// the common "enter today's exam marks for the whole class" workflow.
type BulkRecordMarksRequest struct {
	Results []RecordMarksRequest `json:"results"`
}

// Service implements exam CRUD and marks entry/retrieval.
type Service struct {
	repo *repository.ExamRepository
}

// NewService creates an exam Service.
func NewService(repo *repository.ExamRepository) *Service {
	return &Service{repo: repo}
}

// CreateExam inserts a new exam for a school/class/subject.
func (s *Service) CreateExam(ctx context.Context, schoolID, createdBy string, req CreateExamRequest) (*model.Exam, error) {
	if req.Name == "" || req.ClassID == "" || req.SubjectID == "" {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "name, classId and subjectId are required")
	}
	examType := model.ExamType(req.ExamType)
	switch examType {
	case model.ExamTypeUnitTest, model.ExamTypeMidTerm, model.ExamTypeFinal, model.ExamTypeQuiz:
	default:
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "examType must be one of UNIT_TEST, MID_TERM, FINAL, QUIZ")
	}
	if req.MaxMarks <= 0 {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "maxMarks must be greater than 0")
	}
	examDate, err := time.Parse("2006-01-02", req.ExamDate)
	if err != nil {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "examDate must be in YYYY-MM-DD format")
	}

	exists, err := s.repo.ExistsByClassSubjectTypeDate(ctx, req.ClassID, req.SubjectID, string(examType), examDate)
	if err != nil {
		return nil, fmt.Errorf("exam.CreateExam: check existing: %w", err)
	}
	if exists {
		return nil, apperrors.WithMessage(apperrors.ErrConflict,
			"an exam of this type for this class and subject is already scheduled on this date")
	}

	exam := &model.Exam{
		ID:        uuid.NewString(),
		SchoolID:  schoolID,
		Name:      req.Name,
		ClassID:   req.ClassID,
		SubjectID: req.SubjectID,
		ExamType:  examType,
		ExamDate:  examDate,
		MaxMarks:  req.MaxMarks,
		CreatedBy: createdBy,
	}
	if err := s.repo.CreateExam(ctx, exam); err != nil {
		return nil, fmt.Errorf("exam.CreateExam: %w", err)
	}
	return exam, nil
}

// GetExam fetches an exam by ID, scoped to the caller's school.
func (s *Service) GetExam(ctx context.Context, schoolID, id string) (*model.Exam, error) {
	exam, err := s.repo.GetExamByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("exam.GetExam: %w", err)
	}
	if exam.SchoolID != schoolID {
		return nil, apperrors.ErrNotFound
	}
	return exam, nil
}

// ListExamsByClass lists exams scheduled for a class.
func (s *Service) ListExamsByClass(ctx context.Context, classID string, limit, offset int) ([]model.Exam, error) {
	exams, err := s.repo.ListExamsByClass(ctx, classID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("exam.ListExamsByClass: %w", err)
	}
	return exams, nil
}

// DeleteExam removes an exam by ID.
func (s *Service) DeleteExam(ctx context.Context, schoolID, id string) error {
	if _, err := s.GetExam(ctx, schoolID, id); err != nil {
		return err
	}
	if err := s.repo.DeleteExam(ctx, id); err != nil {
		return fmt.Errorf("exam.DeleteExam: %w", err)
	}
	return nil
}

// RecordMarks records (or corrects) one student's marks for an exam.
func (s *Service) RecordMarks(ctx context.Context, schoolID, examID, evaluatedBy string, req RecordMarksRequest) (*model.ExamResult, error) {
	exam, err := s.GetExam(ctx, schoolID, examID)
	if err != nil {
		return nil, err
	}
	if req.StudentID == "" {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "studentId is required")
	}
	if req.MarksObtained < 0 || req.MarksObtained > exam.MaxMarks {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, fmt.Sprintf("marksObtained must be between 0 and %.2f", exam.MaxMarks))
	}

	result := &model.ExamResult{
		ID:            uuid.NewString(),
		ExamID:        examID,
		StudentID:     req.StudentID,
		MarksObtained: req.MarksObtained,
		Grade:         gradeFor(req.MarksObtained, exam.MaxMarks),
		Remarks:       req.Remarks,
		EvaluatedBy:   evaluatedBy,
	}
	if err := s.repo.UpsertResult(ctx, result); err != nil {
		return nil, fmt.Errorf("exam.RecordMarks: %w", err)
	}
	return result, nil
}

// RecordMarksBulk records marks for multiple students against one exam.
func (s *Service) RecordMarksBulk(ctx context.Context, schoolID, examID, evaluatedBy string, req BulkRecordMarksRequest) ([]model.ExamResult, error) {
	if len(req.Results) == 0 {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "results must not be empty")
	}
	results := make([]model.ExamResult, 0, len(req.Results))
	for _, r := range req.Results {
		res, err := s.RecordMarks(ctx, schoolID, examID, evaluatedBy, r)
		if err != nil {
			return nil, err
		}
		results = append(results, *res)
	}
	return results, nil
}

// ListResultsForExam lists every student's result for an exam.
func (s *Service) ListResultsForExam(ctx context.Context, schoolID, examID string) ([]model.ExamResult, error) {
	if _, err := s.GetExam(ctx, schoolID, examID); err != nil {
		return nil, err
	}
	results, err := s.repo.ListResultsForExam(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam.ListResultsForExam: %w", err)
	}
	return results, nil
}

// Marksheet builds a student's full consolidated marksheet across all exams.
func (s *Service) Marksheet(ctx context.Context, studentID string) (*model.Marksheet, error) {
	entries, err := s.repo.ListResultsForStudent(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("exam.Marksheet: %w", err)
	}
	return &model.Marksheet{StudentID: studentID, Entries: entries}, nil
}

// gradeFor applies a simple percentage-based grading scale. This is a
// reasonable phase-1 default; schools that need custom grading bands can
// request that as a phase-2 configuration.
func gradeFor(marksObtained, maxMarks float64) string {
	if maxMarks <= 0 {
		return ""
	}
	pct := marksObtained / maxMarks * 100
	switch {
	case pct >= 90:
		return "A+"
	case pct >= 80:
		return "A"
	case pct >= 70:
		return "B+"
	case pct >= 60:
		return "B"
	case pct >= 50:
		return "C"
	case pct >= 40:
		return "D"
	default:
		return "F"
	}
}
