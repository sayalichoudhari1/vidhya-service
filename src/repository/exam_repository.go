package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"vidhya-service/src/model"
)

// ExamRepository persists exams and per-student exam results (marksheet).
type ExamRepository struct {
	db *gorm.DB
}

// NewExamRepository creates an ExamRepository.
func NewExamRepository(db *gorm.DB) *ExamRepository {
	return &ExamRepository{db: db}
}

// ------------------------------------------------------------------ exams ---

// CreateExam inserts a new exam.
func (r *ExamRepository) CreateExam(ctx context.Context, e *model.Exam) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return fmt.Errorf("ExamRepository.CreateExam: %w", err)
	}
	return nil
}

// ExistsByClassSubjectTypeDate reports whether an exam of the same type,
// for the same class and subject, is already scheduled on this date -
// catches accidental double-entry of the same exam.
func (r *ExamRepository) ExistsByClassSubjectTypeDate(ctx context.Context, classID, subjectID, examType string, examDate time.Time) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Exam{}).
		Where("class_id = ? AND subject_id = ? AND exam_type = ? AND exam_date = ?", classID, subjectID, examType, examDate).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ExamRepository.ExistsByClassSubjectTypeDate: %w", err)
	}
	return count > 0, nil
}

// GetExamByID fetches an exam by ID.
func (r *ExamRepository) GetExamByID(ctx context.Context, id string) (*model.Exam, error) {
	var e model.Exam
	if err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("ExamRepository.GetExamByID: %w", err)
	}
	return &e, nil
}

// ListExamsByClass lists exams scheduled for a class.
func (r *ExamRepository) ListExamsByClass(ctx context.Context, classID string, limit, offset int) ([]model.Exam, error) {
	var exams []model.Exam
	err := r.db.WithContext(ctx).Where("class_id = ?", classID).
		Order("exam_date desc").Limit(limit).Offset(offset).Find(&exams).Error
	if err != nil {
		return nil, fmt.Errorf("ExamRepository.ListExamsByClass: %w", err)
	}
	return exams, nil
}

// UpdateExam persists changes to an existing exam.
func (r *ExamRepository) UpdateExam(ctx context.Context, e *model.Exam) error {
	if err := r.db.WithContext(ctx).Model(&model.Exam{}).Where("id = ?", e.ID).Updates(e).Error; err != nil {
		return fmt.Errorf("ExamRepository.UpdateExam: %w", err)
	}
	return nil
}

// DeleteExam removes an exam by ID (results cascade via FK, see migration).
func (r *ExamRepository) DeleteExam(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Exam{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("ExamRepository.DeleteExam: %w", err)
	}
	return nil
}

// ------------------------------------------------------------- exam_results ---

// UpsertResult inserts a student's exam result, or updates it if one already
// exists for the same (exam, student) pair.
func (r *ExamRepository) UpsertResult(ctx context.Context, res *model.ExamResult) error {
	var existing model.ExamResult
	err := r.db.WithContext(ctx).Where("exam_id = ? AND student_id = ?", res.ExamID, res.StudentID).First(&existing).Error
	switch {
	case err == nil:
		existing.MarksObtained = res.MarksObtained
		existing.Grade = res.Grade
		existing.Remarks = res.Remarks
		existing.EvaluatedBy = res.EvaluatedBy
		if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return fmt.Errorf("ExamRepository.UpsertResult: update: %w", err)
		}
		*res = existing
		return nil
	case err == gorm.ErrRecordNotFound:
		if err := r.db.WithContext(ctx).Create(res).Error; err != nil {
			return fmt.Errorf("ExamRepository.UpsertResult: create: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("ExamRepository.UpsertResult: lookup: %w", err)
	}
}

// ListResultsForExam lists every student's result for an exam.
func (r *ExamRepository) ListResultsForExam(ctx context.Context, examID string) ([]model.ExamResult, error) {
	var rows []model.ExamResult
	if err := r.db.WithContext(ctx).Where("exam_id = ?", examID).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("ExamRepository.ListResultsForExam: %w", err)
	}
	return rows, nil
}

// ListResultsForStudent lists every exam result recorded for a student,
// joined with exam/subject metadata, to build a consolidated marksheet.
func (r *ExamRepository) ListResultsForStudent(ctx context.Context, studentID string) ([]model.MarksheetEntry, error) {
	var rows []model.MarksheetEntry
	err := r.db.WithContext(ctx).
		Table("exam_results er").
		Select(`e.id as exam_id, e.name as exam_name, e.subject_id, sub.name as subject_name,
				e.exam_type, e.exam_date, e.max_marks, er.marks_obtained, er.grade`).
		Joins("JOIN exams e ON e.id = er.exam_id").
		Joins("JOIN subjects sub ON sub.id = e.subject_id").
		Where("er.student_id = ?", studentID).
		Order("e.exam_date desc").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("ExamRepository.ListResultsForStudent: %w", err)
	}
	return rows, nil
}
