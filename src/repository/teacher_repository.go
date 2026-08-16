package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"vidhya-service/src/model"
)

// TeacherRepository persists the teachers table and joins it with users.
type TeacherRepository struct {
	db *gorm.DB
}

// NewTeacherRepository creates a TeacherRepository.
func NewTeacherRepository(db *gorm.DB) *TeacherRepository {
	return &TeacherRepository{db: db}
}

// Create inserts a new teacher row.
func (r *TeacherRepository) Create(ctx context.Context, t *model.Teacher) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("TeacherRepository.Create: %w", err)
	}
	return nil
}

// ExistsByEmployeeNumber reports whether a teacher with this employee
// number already exists within the school.
func (r *TeacherRepository) ExistsByEmployeeNumber(ctx context.Context, schoolID, employeeNumber string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Teacher{}).
		Where("school_id = ? AND employee_number = ?", schoolID, employeeNumber).Count(&count).Error; err != nil {
		return false, fmt.Errorf("TeacherRepository.ExistsByEmployeeNumber: %w", err)
	}
	return count > 0, nil
}

// GetByUserID fetches a teacher row by user ID.
func (r *TeacherRepository) GetByUserID(ctx context.Context, userID string) (*model.Teacher, error) {
	var t model.Teacher
	if err := r.db.WithContext(ctx).First(&t, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("TeacherRepository.GetByUserID: %w", err)
	}
	return &t, nil
}

// ListBySchool lists teachers for a school, joined with user details.
func (r *TeacherRepository) ListBySchool(ctx context.Context, schoolID string, limit, offset int) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND role = ?", schoolID, model.RoleTeacher).
		Order("first_name, last_name").Limit(limit).Offset(offset).
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("TeacherRepository.ListBySchool: %w", err)
	}
	return users, nil
}

// Delete removes a teacher's role-specific row.
func (r *TeacherRepository) Delete(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Teacher{}, "user_id = ?", userID).Error; err != nil {
		return fmt.Errorf("TeacherRepository.Delete: %w", err)
	}
	return nil
}
