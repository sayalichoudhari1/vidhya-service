package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"vidhya-service/src/model"
)

// UserRepository persists the users table (login identity + RBAC role for
// every principal: admin, principal, teacher, student, parent).
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user row.
func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("UserRepository.Create: %w", err)
	}
	return nil
}

// GetByID fetches a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("UserRepository.GetByID: %w", err)
	}
	return &u, nil
}

// GetByEmail fetches a user by email (case-insensitive), used for login.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, "lower(email) = lower(?)", email).Error; err != nil {
		return nil, fmt.Errorf("UserRepository.GetByEmail: %w", err)
	}
	return &u, nil
}

// Update persists changes to an existing user row.
func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", u.ID).Updates(u).Error; err != nil {
		return fmt.Errorf("UserRepository.Update: %w", err)
	}
	return nil
}

// Delete soft-deletes a user by ID (deactivates rather than hard-deleting, to
// preserve referential integrity with attendance/exam history).
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("UserRepository.Delete: %w", err)
	}
	return nil
}

// ListBySchool lists users for a school, optionally filtered by role.
func (r *UserRepository) ListBySchool(ctx context.Context, schoolID string, role *model.Role, limit, offset int) ([]model.User, error) {
	var users []model.User
	q := r.db.WithContext(ctx).Where("school_id = ?", schoolID)
	if role != nil {
		q = q.Where("role = ?", *role)
	}
	if err := q.Order("created_at desc").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("UserRepository.ListBySchool: %w", err)
	}
	return users, nil
}

// ExistsByEmail reports whether an active user with this email already exists.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("lower(email) = lower(?) AND is_active = true", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("UserRepository.ExistsByEmail: %w", err)
	}
	return count > 0, nil
}

// ExistsByNamePhone reports whether an active user with the same first
// name, last name, and phone number already exists. This is the guard
// against accidentally creating the same person twice (e.g. re-submitting
// a student/teacher onboarding form): names alone are too common to be a
// reliable identity signal, but name+phone together is a strong duplicate
// indicator. schoolID is nil for platform-wide ADMIN accounts, in which
// case the check is scoped to other school-less admins only.
func (r *UserRepository) ExistsByNamePhone(ctx context.Context, schoolID *string, firstName, lastName, phone string) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&model.User{}).
		Where("lower(first_name) = lower(?) AND lower(last_name) = lower(?) AND phone = ? AND is_active = true",
			firstName, lastName, phone)
	if schoolID != nil {
		q = q.Where("school_id = ?", *schoolID)
	} else {
		q = q.Where("school_id IS NULL")
	}
	if err := q.Count(&count).Error; err != nil {
		return false, fmt.Errorf("UserRepository.ExistsByNamePhone: %w", err)
	}
	return count > 0, nil
}
