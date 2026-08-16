package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"vidhya-service/src/model"
)

// StudentRepository persists the students table and joins it with users for
// read-friendly responses.
type StudentRepository struct {
	db *gorm.DB
}

// NewStudentRepository creates a StudentRepository.
func NewStudentRepository(db *gorm.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

// Create inserts a new student enrollment row.
func (r *StudentRepository) Create(ctx context.Context, s *model.Student) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return fmt.Errorf("StudentRepository.Create: %w", err)
	}
	return nil
}

// GetByUserID fetches the student enrollment row joined with user details.
func (r *StudentRepository) GetByUserID(ctx context.Context, userID string) (*model.StudentWithUser, error) {
	var out model.StudentWithUser
	err := r.db.WithContext(ctx).
		Table("students s").
		Select(`s.*, u.first_name, u.last_name, u.email, u.phone, u.is_active`).
		Joins("JOIN users u ON u.id = s.user_id").
		Where("s.user_id = ?", userID).
		Scan(&out).Error
	if err != nil {
		return nil, fmt.Errorf("StudentRepository.GetByUserID: %w", err)
	}
	if out.UserID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &out, nil
}

// Update persists changes to an existing student enrollment row.
func (r *StudentRepository) Update(ctx context.Context, s *model.Student) error {
	if err := r.db.WithContext(ctx).Model(&model.Student{}).Where("user_id = ?", s.UserID).Updates(s).Error; err != nil {
		return fmt.Errorf("StudentRepository.Update: %w", err)
	}
	return nil
}

// Delete removes a student's enrollment row. The underlying user record is
// deactivated separately by the service layer via UserRepository.Delete.
func (r *StudentRepository) Delete(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Student{}, "user_id = ?", userID).Error; err != nil {
		return fmt.Errorf("StudentRepository.Delete: %w", err)
	}
	return nil
}

// ListBySchool lists students for a school, optionally filtered by class.
func (r *StudentRepository) ListBySchool(ctx context.Context, schoolID string, classID *string, limit, offset int) ([]model.StudentWithUser, error) {
	var out []model.StudentWithUser
	q := r.db.WithContext(ctx).
		Table("students s").
		Select(`s.*, u.first_name, u.last_name, u.email, u.phone, u.is_active`).
		Joins("JOIN users u ON u.id = s.user_id").
		Where("s.school_id = ?", schoolID)
	if classID != nil {
		q = q.Where("s.class_id = ?", *classID)
	}
	if err := q.Order("u.first_name, u.last_name").Limit(limit).Offset(offset).Scan(&out).Error; err != nil {
		return nil, fmt.Errorf("StudentRepository.ListBySchool: %w", err)
	}
	return out, nil
}

// ExistsByAdmissionNumber reports whether a student with this admission
// number already exists within the school.
func (r *StudentRepository) ExistsByAdmissionNumber(ctx context.Context, schoolID, admissionNumber string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Student{}).
		Where("school_id = ? AND admission_number = ?", schoolID, admissionNumber).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("StudentRepository.ExistsByAdmissionNumber: %w", err)
	}
	return count > 0, nil
}

// ExistsByClassRollNumber reports whether another student already holds
// this roll number within the same class - two students in one class
// cannot share a roll number. excludeUserID is skipped when checking (used
// on update, so a student doesn't collide with their own current row).
func (r *StudentRepository) ExistsByClassRollNumber(ctx context.Context, classID, rollNumber, excludeUserID string) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&model.Student{}).
		Where("class_id = ? AND roll_number = ?", classID, rollNumber)
	if excludeUserID != "" {
		q = q.Where("user_id <> ?", excludeUserID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, fmt.Errorf("StudentRepository.ExistsByClassRollNumber: %w", err)
	}
	return count > 0, nil
}
